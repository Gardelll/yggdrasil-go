/*
 * Copyright (C) 2022-2025. Gardel <gardel741@outlook.com> and contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
	"yggdrasil-go/dto"
	"yggdrasil-go/util"
)

// UpstreamStatus represents the status of an upstream service
type UpstreamStatus string

const (
	StatusAvailable   UpstreamStatus = "available"
	StatusUnavailable UpstreamStatus = "unavailable"
)

// AggregationStrategy defines the response aggregation strategy
type AggregationStrategy string

const (
	RaceToSuccess AggregationStrategy = "race_to_success" // Return first success
	WaitAll       AggregationStrategy = "wait_all"        // Wait for all responses
)

const (
	OpLookupProfile = "LookupProfile"
	OpLookupByName  = "LookupByName"
	OpLookupByUUID  = "LookupByUUID"
	OpLookupBulk    = "LookupBulkProfiles"
	OpHasJoined     = "HasJoined"
	OpGetPublicKeys = "GetPublicKeys"
	OpVerifySession = "VerifySession"
)

// UpstreamServiceState represents the runtime state of an upstream service
type UpstreamServiceState struct {
	ID              string        // Unique identifier
	ProfileURL      string        // Session profile query endpoint (supports {uuid} placeholder)
	LookupByNameURL string        // Lookup by username endpoint (supports {username} placeholder)
	LookupByUUIDURL string        // Lookup by UUID endpoint (supports {uuid} placeholder)
	BulkLookupURL   string        // Bulk lookup endpoint (POST)
	JoinURL         string        // Join server endpoint (POST)
	HasJoinedURL    string        // Verify has joined endpoint (supports query parameters)
	PublicKeysURL   string        // Public keys endpoint
	Timeout         time.Duration // Request timeout
	Status          UpstreamStatus
	LastErr         error
	LastFail        time.Time
	FailCnt         int
	RetryAt         time.Time

	// Statistics
	TotalRequests int64
	SuccessCount  int64
	TotalDuration time.Duration

	mu sync.RWMutex // Protects the state fields
}

// UpstreamRequest represents a request to upstream services
type UpstreamRequest struct {
	ID          string
	Timestamp   time.Time
	Operation   string
	AggStrategy AggregationStrategy
	Deadline    time.Time
}

// UpstreamResponse represents a response from an upstream service
type UpstreamResponse struct {
	RequestID  string
	UpstreamID string
	Timestamp  time.Time
	StatusCode int
	Body       json.RawMessage
	Duration   time.Duration
	IsSuccess  bool
	Error      error
	ErrorMsg   string
}

// AggregatedResult represents the aggregated result from multiple upstreams
type AggregatedResult struct {
	RequestID      string
	Strategy       AggregationStrategy
	Timestamp      time.Time
	TotalUpstreams int
	SuccessCount   int
	FailureCount   int
	PrimaryResult  *UpstreamResponse
	AllResults     []*UpstreamResponse
	Consistency    bool
	IsSuccess      bool
	Error          error
	ErrorDetails   map[string]error
}

// UpstreamServicePool manages the goroutine pool for concurrent requests
type UpstreamServicePool struct {
	MaxSize   int
	semaphore *semaphore.Weighted

	// Statistics
	ActiveCount   int
	TotalRequests int64
	TotalDuration time.Duration

	mu sync.Mutex
}

// NewUpstreamServicePool creates a new pool with the given size
func NewUpstreamServicePool(size int) *UpstreamServicePool {
	return &UpstreamServicePool{
		MaxSize:   size,
		semaphore: semaphore.NewWeighted(int64(size)),
	}
}

// Acquire acquires a permit from the pool
func (p *UpstreamServicePool) Acquire(ctx context.Context) error {
	err := p.semaphore.Acquire(ctx, 1)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.ActiveCount++
	p.TotalRequests++
	p.mu.Unlock()

	return nil
}

// Release releases a permit back to the pool
func (p *UpstreamServicePool) Release() {
	p.mu.Lock()
	p.ActiveCount--
	p.mu.Unlock()

	p.semaphore.Release(1)
}

// IUpstreamService defines the upstream authentication service interface
type IUpstreamService interface {
	// LookupProfile queries a single player profile by UUID
	LookupProfile(ctx context.Context, profileID string, unsigned bool) (*dto.UpstreamProfileResponse, error)

	// LookupByName queries a single player profile by username (tasks.md T017)
	LookupByName(ctx context.Context, username string) (*dto.UpstreamProfileResponse, error)

	// LookupByUUID queries a single player profile by UUID (tasks.md T017)
	LookupByUUID(ctx context.Context, uuid string) (*dto.UpstreamProfileResponse, error)

	// LookupBulkProfiles queries multiple player profiles by username
	LookupBulkProfiles(ctx context.Context, usernames []string) ([]*dto.ProfileResponse, error)

	// VerifySession verifies a player session (join server)
	VerifySession(ctx context.Context, accessToken, selectedProfile, serverId string) error

	// HasJoined checks if a player has joined a server
	HasJoined(ctx context.Context, username, serverId string, ipAddress *string) (*dto.JoinedResponse, error)

	// GetPublicKeys retrieves the public keys from upstream
	GetPublicKeys(ctx context.Context) (*dto.PublicKeysResponse, error)

	// GetUpstreamStatus returns the status of all upstream services (for monitoring)
	GetUpstreamStatus() map[string]*UpstreamStatusInfo
}

// Response data structures moved to dto package

type UpstreamStatusInfo struct {
	ID                string
	URL               string
	Status            string
	LastFailure       *time.Time
	LastFailureReason string
	FailureCount      int
	TotalRequests     int64
	SuccessCount      int64
	AverageLatency    time.Duration
}

// Error types

var (
	ErrAllUpstreamsFailed = errors.New("all upstream services failed")
	ErrUpstreamTimeout    = errors.New("upstream request timeout")
	ErrInvalidResponse    = errors.New("invalid upstream response")
)

// upstreamService is the implementation of IUpstreamService
type upstreamService struct {
	config       *util.UpstreamConfig
	pool         *UpstreamServicePool
	upstreams    map[string]*UpstreamServiceState
	client       *http.Client
	degradedMode bool // Degraded mode flag (no upstream services configured)

	// Request counter for retry mechanism
	requestCounter int64
	counterMu      sync.Mutex
}

// NewUpstreamService creates a new upstream service instance
func NewUpstreamService(config *util.UpstreamConfig, upstreamConfigs []*util.UpstreamServiceConfig) (IUpstreamService, error) {
	if config == nil {
		return nil, errors.New("upstream config cannot be nil")
	}

	// Initialize the goroutine pool
	pool := NewUpstreamServicePool(config.PoolSize)

	// Create HTTP client with reasonable defaults
	client := &http.Client{
		Timeout: 30 * time.Second, // Global timeout, individual requests use their own
	}

	// Detect degraded mode
	degradedMode := len(upstreamConfigs) == 0

	if degradedMode {
		log.Println("警告，未配置上游认证服务，将使用降级模式运行（纯本地认证）")
	}

	// Initialize upstream states
	upstreams := make(map[string]*UpstreamServiceState)

	for _, cfg := range upstreamConfigs {
		if cfg == nil {
			continue
		}

		state := &UpstreamServiceState{
			ID:              cfg.Id,
			ProfileURL:      cfg.ProfileURL,
			LookupByNameURL: cfg.LookupByNameURL,
			LookupByUUIDURL: cfg.LookupByUUIDURL,
			BulkLookupURL:   cfg.BulkLookupURL,
			JoinURL:         cfg.JoinURL,
			HasJoinedURL:    cfg.HasJoinedURL,
			PublicKeysURL:   cfg.PublicKeysURL,
			Timeout:         time.Duration(cfg.Timeout) * time.Millisecond,
			Status:          StatusAvailable, // Initially all upstreams are available
		}

		upstreams[cfg.Id] = state
	}

	if !degradedMode {
		log.Printf("已初始化 %d 个上游验证服务: %v", len(upstreams), config.Services)
		for id, state := range upstreams {
			log.Printf("  - %s: profile=%s, byname=%s, byuuid=%s, bulk=%s, join=%s, hasJoined=%s, publicKeys=%s",
				id, state.ProfileURL, state.LookupByNameURL, state.LookupByUUIDURL, state.BulkLookupURL, state.JoinURL, state.HasJoinedURL, state.PublicKeysURL)
		}
	}

	svc := &upstreamService{
		config:       config,
		pool:         pool,
		upstreams:    upstreams,
		client:       client,
		degradedMode: degradedMode,
	}

	return svc, nil
}

// GetUpstreamStatus returns the current status of all upstream services
func (u *upstreamService) GetUpstreamStatus() map[string]*UpstreamStatusInfo {
	result := make(map[string]*UpstreamStatusInfo)

	for id, state := range u.upstreams {
		state.mu.RLock()
		info := &UpstreamStatusInfo{
			ID:            state.ID,
			URL:           state.ProfileURL, // Use ProfileURL as representative
			Status:        string(state.Status),
			FailureCount:  state.FailCnt,
			TotalRequests: state.TotalRequests,
			SuccessCount:  state.SuccessCount,
		}

		if !state.LastFail.IsZero() {
			info.LastFailure = &state.LastFail
		}

		if state.LastErr != nil {
			info.LastFailureReason = state.LastErr.Error()
		}

		if state.SuccessCount > 0 {
			info.AverageLatency = time.Duration(state.TotalDuration.Nanoseconds() / state.SuccessCount)
		}

		state.mu.RUnlock()
		result[id] = info
	}

	return result
}

// markUpstreamUnavailable marks an upstream as unavailable
func (u *upstreamService) markUpstreamUnavailable(upstreamID string, err error) {
	state, exists := u.upstreams[upstreamID]
	if !exists {
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	state.Status = StatusUnavailable
	state.LastErr = err
	state.LastFail = time.Now()
	state.FailCnt++

	// Set retry time based on recovery timeout
	state.RetryAt = time.Now().Add(time.Duration(u.config.RecoveryTimeout) * time.Millisecond)

	log.Printf("Upstream %s marked as unavailable: %v (failure count: %d)", upstreamID, err, state.FailCnt)
}

// markUpstreamAvailable marks an upstream as available (after successful recovery)
func (u *upstreamService) markUpstreamAvailable(upstreamID string) {
	state, exists := u.upstreams[upstreamID]
	if !exists {
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.Status == StatusUnavailable {
		log.Printf("Upstream %s recovered and marked as available", upstreamID)
	}

	state.Status = StatusAvailable
	state.LastErr = nil
}

// shouldRetryUnavailableUpstream determines if an unavailable upstream should be retried
func (u *upstreamService) shouldRetryUnavailableUpstream(state *UpstreamServiceState) bool {
	state.mu.RLock()
	defer state.mu.RUnlock()

	if state.Status == StatusAvailable {
		return true
	}

	// Check if recovery timeout has passed
	if time.Now().After(state.RetryAt) {
		return true
	}

	// Check if retry interval (number of requests) has passed
	u.counterMu.Lock()
	shouldRetry := u.requestCounter%int64(u.config.RetryInterval) == 0
	u.counterMu.Unlock()

	return shouldRetry
}

// incrementRequestCounter increments the global request counter
func (u *upstreamService) incrementRequestCounter() {
	u.counterMu.Lock()
	u.requestCounter++
	u.counterMu.Unlock()
}

// getAvailableUpstreams returns a list of available upstream services
func (u *upstreamService) getAvailableUpstreams() []*UpstreamServiceState {
	var available []*UpstreamServiceState

	for _, state := range u.upstreams {
		state.mu.RLock()
		isAvailable := state.Status == StatusAvailable
		state.mu.RUnlock()

		if isAvailable {
			available = append(available, state)
		} else if u.shouldRetryUnavailableUpstream(state) {
			// Include unavailable upstreams that should be retried
			available = append(available, state)
		}
	}

	return available
}

// generateRequestID generates a unique request ID for tracking
func generateRequestID() string {
	var buf [12]byte
	_, _ = rand.Read(buf[:])
	return fmt.Sprintf("req-%s", hex.EncodeToString(buf[:]))
}

// logRequest logs the request details
func (u *upstreamService) logRequest(req *UpstreamRequest, result *AggregatedResult) {
	if result.IsSuccess {
		log.Printf("[Request %s] %s completed successfully in %v from %s (attempted: %d, success: %d)",
			req.ID, req.Operation, time.Since(req.Timestamp),
			result.PrimaryResult.UpstreamID, result.TotalUpstreams, result.SuccessCount)
	} else {
		log.Printf("[Request %s] %s FAILED after %v (attempted: %d, all failed)",
			req.ID, req.Operation, time.Since(req.Timestamp), result.TotalUpstreams)
		for upstreamID, err := range result.ErrorDetails {
			log.Printf("  - %s: %v", upstreamID, err)
		}
	}
}

// requestAllUpstreams sends concurrent requests to all available upstreams
func (u *upstreamService) requestAllUpstreams(ctx context.Context, req *UpstreamRequest,
	reqFunc func(*UpstreamServiceState) (*UpstreamResponse, error)) *AggregatedResult {

	availableUpstreams := u.getAvailableUpstreams()
	if len(availableUpstreams) == 0 {
		return &AggregatedResult{
			RequestID:    req.ID,
			Strategy:     req.AggStrategy,
			Timestamp:    time.Now(),
			IsSuccess:    false,
			Error:        ErrAllUpstreamsFailed,
			ErrorDetails: map[string]error{"all": errors.New("no upstream services available")},
		}
	}

	// Channel to collect responses
	responsesChan := make(chan *UpstreamResponse, len(availableUpstreams))
	ctx, cancel := context.WithDeadline(ctx, req.Deadline)
	defer cancel()

	// Launch goroutines for each upstream
	var wg sync.WaitGroup
	for _, upstream := range availableUpstreams {
		wg.Add(1)
		go func(state *UpstreamServiceState) {
			defer wg.Done()

			start := time.Now()
			resp, err := reqFunc(state)

			duration := time.Since(start)

			// Build response
			upstreamResp := &UpstreamResponse{
				RequestID:  req.ID,
				UpstreamID: state.ID,
				Timestamp:  time.Now(),
				Duration:   duration,
			}

			if err != nil {
				upstreamResp.IsSuccess = false
				upstreamResp.Error = err
				upstreamResp.ErrorMsg = err.Error()

				// Mark upstream as unavailable
				u.markUpstreamUnavailable(state.ID, err)
			} else if resp != nil {
				upstreamResp.IsSuccess = resp.IsSuccess
				upstreamResp.StatusCode = resp.StatusCode
				upstreamResp.Body = resp.Body
				upstreamResp.Error = resp.Error

				// Update statistics
				state.mu.Lock()
				state.SuccessCount++
				state.TotalRequests++
				state.TotalDuration += duration
				state.mu.Unlock()

				// Mark as available (in case it was recovered)
				u.markUpstreamAvailable(state.ID)
			}

			select {
			case responsesChan <- upstreamResp:
			case <-ctx.Done():
				// Context cancelled, don't send response
			}
		}(upstream)
	}

	// Close channel after all goroutines complete
	go func() {
		wg.Wait()
		close(responsesChan)
	}()

	// Collect responses based on strategy
	return u.aggregateResponses(req, responsesChan, len(availableUpstreams))
}

// aggregateResponses aggregates responses based on the strategy
func (u *upstreamService) aggregateResponses(req *UpstreamRequest, responsesChan <-chan *UpstreamResponse, totalUpstreams int) *AggregatedResult {

	result := &AggregatedResult{
		RequestID:      req.ID,
		Strategy:       req.AggStrategy,
		Timestamp:      time.Now(),
		TotalUpstreams: totalUpstreams,
		AllResults:     make([]*UpstreamResponse, 0, totalUpstreams),
		ErrorDetails:   make(map[string]error),
	}

	if req.AggStrategy == RaceToSuccess {
		// Return first successful response
		for resp := range responsesChan {
			result.AllResults = append(result.AllResults, resp)

			if resp.IsSuccess {
				result.PrimaryResult = resp
				result.SuccessCount++
				result.IsSuccess = true
				// We have a success, we're done
				return result
			} else {
				result.FailureCount++
				if resp.Error != nil {
					result.ErrorDetails[resp.UpstreamID] = resp.Error
				}
			}
		}

		// No successful response
		result.IsSuccess = false
		result.Error = ErrAllUpstreamsFailed
		return result

	} else if req.AggStrategy == WaitAll {
		// Wait for all responses
		for resp := range responsesChan {
			result.AllResults = append(result.AllResults, resp)

			if resp.IsSuccess {
				result.SuccessCount++
				if result.PrimaryResult == nil {
					result.PrimaryResult = resp
				}
			} else {
				result.FailureCount++
				if resp.Error != nil {
					result.ErrorDetails[resp.UpstreamID] = resp.Error
				}
			}
		}

		result.IsSuccess = result.SuccessCount > 0
		if !result.IsSuccess {
			result.Error = ErrAllUpstreamsFailed
		}

		return result
	}

	// Unknown strategy
	result.IsSuccess = false
	result.Error = errors.New("unknown aggregation strategy")
	return result
}

// doHTTPRequestWithFullURL performs an HTTP request to a specific upstream with full URL
// treat204AsSuccess: true if 204 means success (for operations like join), false if 204 means not found (for queries)
func (u *upstreamService) doHTTPRequestWithFullURL(ctx context.Context, upstream *UpstreamServiceState,
	method, fullURL string, body []byte, treat204AsSuccess bool) (*UpstreamResponse, error) {

	// Use the generalized HTTP utility function
	httpResp, err := util.DoHTTPRequestWithContext(ctx, u.client, method, fullURL, body, upstream.Timeout)

	if err != nil && httpResp == nil {
		return nil, err
	}

	// Handle 204 No Content based on operation type (per Yggdrasil spec)
	// For query operations (LookupProfile, HasJoined, etc.), 204 means "not found" (failure)
	// For operation endpoints (VerifySession, etc.), 204 means "success"
	isSuccess := httpResp.IsSuccess
	if httpResp.StatusCode == http.StatusNoContent {
		isSuccess = treat204AsSuccess
	}

	// Convert util.HTTPResponse to UpstreamResponse
	upstreamResp := &UpstreamResponse{
		UpstreamID: upstream.ID,
		Timestamp:  time.Now(),
		StatusCode: httpResp.StatusCode,
		Body:       httpResp.Body,
		Duration:   httpResp.Duration,
		IsSuccess:  isSuccess,
		Error:      httpResp.Error,
	}

	if httpResp.Error != nil {
		upstreamResp.ErrorMsg = httpResp.Error.Error()
	}

	return upstreamResp, nil
}

// LookupProfile queries a single player profile
func (u *upstreamService) LookupProfile(ctx context.Context, profileID string, unsigned bool) (*dto.UpstreamProfileResponse, error) {
	// Degraded mode check
	if u.degradedMode {
		log.Printf("Degraded mode: skipping upstream profile lookup for %s", profileID)
		return nil, nil
	}

	u.incrementRequestCounter()

	req := &UpstreamRequest{
		ID:          generateRequestID(),
		Timestamp:   time.Now(),
		Operation:   OpLookupProfile,
		AggStrategy: RaceToSuccess,
		Deadline:    time.Now().Add(30 * time.Second),
	}

	// Acquire pool permit
	if err := u.pool.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire pool permit: %w", err)
	}
	defer u.pool.Release()

	// Send concurrent requests
	result := u.requestAllUpstreams(ctx, req, func(upstream *UpstreamServiceState) (*UpstreamResponse, error) {
		// Use configured ProfileURL with placeholder replacement
		actualURL := util.ReplaceURLPlaceholders(upstream.ProfileURL, map[string]string{
			"uuid": profileID,
		})
		// Add unsigned query parameter
		if unsigned {
			actualURL = fmt.Sprintf("%s?unsigned=true", actualURL)
		} else {
			actualURL = fmt.Sprintf("%s?unsigned=false", actualURL)
		}
		return u.doHTTPRequestWithFullURL(ctx, upstream, "GET", actualURL, nil, false)
	})

	// Log the request
	u.logRequest(req, result)

	if !result.IsSuccess {
		// Check if all upstreams returned 204 (not found)
		// In this case, it's not an error, just means the profile doesn't exist
		all204 := true
		for _, resp := range result.AllResults {
			if resp.StatusCode != http.StatusNoContent {
				all204 = false
				break
			}
		}
		if all204 && len(result.AllResults) > 0 {
			return nil, nil
		}
		return nil, result.Error
	}

	// Parse response
	var profile dto.UpstreamProfileResponse
	if err := json.Unmarshal(result.PrimaryResult.Body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	return &profile, nil
}

// LookupBulkProfiles queries multiple player profiles by username
func (u *upstreamService) LookupBulkProfiles(ctx context.Context, usernames []string) ([]*dto.ProfileResponse, error) {
	// Degraded mode check
	if u.degradedMode {
		log.Printf("Degraded mode: skipping upstream bulk profile lookup")
		return nil, nil
	}

	u.incrementRequestCounter()

	req := &UpstreamRequest{
		ID:          generateRequestID(),
		Timestamp:   time.Now(),
		Operation:   OpLookupBulk,
		AggStrategy: WaitAll, // Wait for all upstreams to ensure we get all possible results
		Deadline:    time.Now().Add(30 * time.Second),
	}

	// Acquire pool permit
	if err := u.pool.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire pool permit: %w", err)
	}
	defer u.pool.Release()

	// Prepare request body
	bodyBytes, err := json.Marshal(usernames)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal usernames: %w", err)
	}

	// Send concurrent requests
	result := u.requestAllUpstreams(ctx, req, func(upstream *UpstreamServiceState) (*UpstreamResponse, error) {
		// Use configured BulkLookupURL (no placeholders needed for bulk lookup)
		return u.doHTTPRequestWithFullURL(ctx, upstream, "POST", upstream.BulkLookupURL, bodyBytes, false)
	})

	// Log the request
	u.logRequest(req, result)

	if !result.IsSuccess {
		return nil, result.Error
	}

	// Parse and merge results from all upstreams
	profileMap := make(map[string]*dto.ProfileResponse)
	for _, resp := range result.AllResults {
		if !resp.IsSuccess {
			continue
		}

		var profiles []*dto.ProfileResponse
		if err := json.Unmarshal(resp.Body, &profiles); err != nil {
			log.Printf("Warning: failed to parse bulk profiles from %s: %v", resp.UpstreamID, err)
			continue
		}

		// Merge profiles (prefer first occurrence)
		for _, profile := range profiles {
			if _, exists := profileMap[profile.Id]; !exists {
				profileMap[profile.Id] = profile
			}
		}
	}

	// Convert map to slice
	profiles := make([]*dto.ProfileResponse, 0, len(profileMap))
	for _, profile := range profileMap {
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// VerifySession verifies a player session (join server)
func (u *upstreamService) VerifySession(ctx context.Context, accessToken, selectedProfile, serverId string) error {
	// Degraded mode check
	if u.degradedMode {
		log.Printf("Degraded mode: skipping upstream session verification")
		return nil
	}

	u.incrementRequestCounter()

	req := &UpstreamRequest{
		ID:          generateRequestID(),
		Timestamp:   time.Now(),
		Operation:   OpVerifySession,
		AggStrategy: RaceToSuccess,
		Deadline:    time.Now().Add(30 * time.Second),
	}

	// Acquire pool permit
	if err := u.pool.Acquire(ctx); err != nil {
		return fmt.Errorf("failed to acquire pool permit: %w", err)
	}
	defer u.pool.Release()

	// Prepare request body - forward all parameters from client request
	reqBody := map[string]string{
		"accessToken":     accessToken,
		"selectedProfile": selectedProfile,
		"serverId":        serverId,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Send concurrent requests
	result := u.requestAllUpstreams(ctx, req, func(upstream *UpstreamServiceState) (*UpstreamResponse, error) {
		// Use configured JoinURL (no placeholders needed)
		return u.doHTTPRequestWithFullURL(ctx, upstream, "POST", upstream.JoinURL, bodyBytes, true)
	})

	// Log the request
	u.logRequest(req, result)

	if !result.IsSuccess {
		return result.Error
	}

	return nil
}

// HasJoined checks if a player has joined a server
func (u *upstreamService) HasJoined(ctx context.Context, username, serverId string, ipAddress *string) (*dto.JoinedResponse, error) {
	// Degraded mode check
	if u.degradedMode {
		log.Printf("Degraded mode: skipping upstream hasJoined verification for %s", username)
		return nil, nil
	}

	u.incrementRequestCounter()

	req := &UpstreamRequest{
		ID:          generateRequestID(),
		Timestamp:   time.Now(),
		Operation:   OpHasJoined,
		AggStrategy: RaceToSuccess,
		Deadline:    time.Now().Add(30 * time.Second),
	}

	// Acquire pool permit
	if err := u.pool.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire pool permit: %w", err)
	}
	defer u.pool.Release()

	// Send concurrent requests
	result := u.requestAllUpstreams(ctx, req, func(upstream *UpstreamServiceState) (*UpstreamResponse, error) {
		// Build full URL with query parameters
		fullURL := fmt.Sprintf("%s?username=%s&serverId=%s", upstream.HasJoinedURL, username, serverId)
		if ipAddress != nil && *ipAddress != "" {
			fullURL += fmt.Sprintf("&ip=%s", *ipAddress)
		}
		return u.doHTTPRequestWithFullURL(ctx, upstream, "GET", fullURL, nil, false)
	})

	// Log the request
	u.logRequest(req, result)

	if !result.IsSuccess {
		// Check if all upstreams returned 204 (verification failed / not found)
		// In this case, it's not an error, just means the session wasn't found
		all204 := true
		for _, resp := range result.AllResults {
			if resp.StatusCode != http.StatusNoContent {
				all204 = false
				break
			}
		}
		if all204 && len(result.AllResults) > 0 {
			return nil, nil
		}
		return nil, result.Error
	}

	// Parse response
	var joinedResp dto.JoinedResponse
	if err := json.Unmarshal(result.PrimaryResult.Body, &joinedResp); err != nil {
		return nil, fmt.Errorf("failed to parse joined response: %w", err)
	}

	return &joinedResp, nil
}

// GetPublicKeys retrieves the public keys from upstream
// Strategy: WaitAll to collect and merge public keys from all upstreams (supports multiple different auth services)
func (u *upstreamService) GetPublicKeys(ctx context.Context) (*dto.PublicKeysResponse, error) {
	// Partial degraded mode: public keys MUST come from upstream
	if u.degradedMode {
		return nil, errors.New("public keys service unavailable: no upstream services configured")
	}

	u.incrementRequestCounter()

	req := &UpstreamRequest{
		ID:          generateRequestID(),
		Timestamp:   time.Now(),
		Operation:   OpGetPublicKeys,
		AggStrategy: WaitAll, // Wait for all upstreams to collect all public keys
		Deadline:    time.Now().Add(30 * time.Second),
	}

	// Acquire pool permit
	if err := u.pool.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire pool permit: %w", err)
	}
	defer u.pool.Release()

	// Send concurrent requests
	result := u.requestAllUpstreams(ctx, req, func(upstream *UpstreamServiceState) (*UpstreamResponse, error) {
		// Use configured PublicKeysURL (no placeholders needed)
		return u.doHTTPRequestWithFullURL(ctx, upstream, "GET", upstream.PublicKeysURL, nil, false)
	})

	// Log the request
	u.logRequest(req, result)

	if !result.IsSuccess {
		return nil, result.Error
	}

	// Merge public keys from all successful upstreams
	mergedKeys := &dto.PublicKeysResponse{
		ProfilePropertyKeys:   make([]dto.PublicKeyItem, 0),
		PlayerCertificateKeys: make([]dto.PublicKeyItem, 0),
	}

	profileKeyMap := make(map[string]dto.PublicKeyItem) // Deduplicate by public key string
	playerKeyMap := make(map[string]dto.PublicKeyItem)

	for _, resp := range result.AllResults {
		if !resp.IsSuccess {
			continue
		}

		var keysResp dto.PublicKeysResponse
		if err := json.Unmarshal(resp.Body, &keysResp); err != nil {
			log.Printf("Warning: failed to parse public keys from %s: %v", resp.UpstreamID, err)
			continue
		}

		// Merge profile property keys (deduplicate by public key value)
		for _, key := range keysResp.ProfilePropertyKeys {
			if _, exists := profileKeyMap[key.PublicKey]; !exists {
				profileKeyMap[key.PublicKey] = key
			}
		}

		// Merge player certificate keys (deduplicate by public key value)
		for _, key := range keysResp.PlayerCertificateKeys {
			if _, exists := playerKeyMap[key.PublicKey]; !exists {
				playerKeyMap[key.PublicKey] = key
			}
		}
	}

	// Convert maps to slices
	for _, key := range profileKeyMap {
		mergedKeys.ProfilePropertyKeys = append(mergedKeys.ProfilePropertyKeys, key)
	}
	for _, key := range playerKeyMap {
		mergedKeys.PlayerCertificateKeys = append(mergedKeys.PlayerCertificateKeys, key)
	}

	if len(mergedKeys.ProfilePropertyKeys) == 0 && len(mergedKeys.PlayerCertificateKeys) == 0 {
		return nil, errors.New("no public keys retrieved from any upstream")
	}

	return mergedKeys, nil
}

// LookupByName queries a single player profile by username
func (u *upstreamService) LookupByName(ctx context.Context, username string) (*dto.UpstreamProfileResponse, error) {
	// Degraded mode check
	if u.degradedMode {
		log.Printf("Degraded mode: skipping upstream profile lookup by name for %s", username)
		return nil, nil
	}

	u.incrementRequestCounter()

	req := &UpstreamRequest{
		ID:          generateRequestID(),
		Timestamp:   time.Now(),
		Operation:   OpLookupByName,
		AggStrategy: RaceToSuccess,
		Deadline:    time.Now().Add(30 * time.Second),
	}

	// Acquire pool permit
	if err := u.pool.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire pool permit: %w", err)
	}
	defer u.pool.Release()

	// Send concurrent requests
	result := u.requestAllUpstreams(ctx, req, func(upstream *UpstreamServiceState) (*UpstreamResponse, error) {
		// Use configured LookupByNameURL with placeholder replacement
		actualURL := util.ReplaceURLPlaceholders(upstream.LookupByNameURL, map[string]string{
			"username": username,
		})
		return u.doHTTPRequestWithFullURL(ctx, upstream, "GET", actualURL, nil, false)
	})

	// Log the request
	u.logRequest(req, result)

	if !result.IsSuccess {
		// Check if all upstreams returned 204 (not found)
		// In this case, it's not an error, just means the profile doesn't exist
		all204 := true
		for _, resp := range result.AllResults {
			if resp.StatusCode != http.StatusNoContent {
				all204 = false
				break
			}
		}
		if all204 && len(result.AllResults) > 0 {
			return nil, nil
		}
		return nil, result.Error
	}

	// Parse response
	var profile dto.UpstreamProfileResponse
	if err := json.Unmarshal(result.PrimaryResult.Body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	return &profile, nil
}

// LookupByUUID queries a single player profile by UUID
func (u *upstreamService) LookupByUUID(ctx context.Context, uuid string) (*dto.UpstreamProfileResponse, error) {
	// Degraded mode check
	if u.degradedMode {
		log.Printf("Degraded mode: skipping upstream profile lookup by UUID for %s", uuid)
		return nil, nil
	}

	u.incrementRequestCounter()

	req := &UpstreamRequest{
		ID:          generateRequestID(),
		Timestamp:   time.Now(),
		Operation:   OpLookupByUUID,
		AggStrategy: RaceToSuccess,
		Deadline:    time.Now().Add(30 * time.Second),
	}

	// Acquire pool permit
	if err := u.pool.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire pool permit: %w", err)
	}
	defer u.pool.Release()

	// Send concurrent requests
	result := u.requestAllUpstreams(ctx, req, func(upstream *UpstreamServiceState) (*UpstreamResponse, error) {
		// Use configured LookupByUUIDURL with placeholder replacement
		actualURL := util.ReplaceURLPlaceholders(upstream.LookupByUUIDURL, map[string]string{
			"uuid": uuid,
		})
		return u.doHTTPRequestWithFullURL(ctx, upstream, "GET", actualURL, nil, false)
	})

	// Log the request
	u.logRequest(req, result)

	if !result.IsSuccess {
		// Check if all upstreams returned 204 (not found)
		// In this case, it's not an error, just means the profile doesn't exist
		all204 := true
		for _, resp := range result.AllResults {
			if resp.StatusCode != http.StatusNoContent {
				all204 = false
				break
			}
		}
		if all204 && len(result.AllResults) > 0 {
			return nil, nil
		}
		return nil, result.Error
	}

	// Parse response
	var profile dto.UpstreamProfileResponse
	if err := json.Unmarshal(result.PrimaryResult.Body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	return &profile, nil
}
