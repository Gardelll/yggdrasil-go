/*
 * Copyright (C) 2022-2025. Gardel <sunxinao@hotmail.com> and contributors
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

package util

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPResponse represents a generic HTTP response with raw body
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Duration   time.Duration
	IsSuccess  bool
	Error      error
}

// DoHTTPRequestWithContext performs an HTTP request with context and timeout support
func DoHTTPRequestWithContext(ctx context.Context, client *http.Client, method, url string, body []byte, timeout time.Duration) (*HTTPResponse, error) {
	// Create request context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create HTTP request
	var req *http.Request
	var err error
	if body != nil && len(body) > 0 {
		req, err = http.NewRequestWithContext(reqCtx, method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(reqCtx, method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if method == "POST" || method == "PUT" || method == "PATCH" {
		// Set headers
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return &HTTPResponse{
			Duration:  duration,
			IsSuccess: false,
			Error:     fmt.Errorf("request failed: %w", err),
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{
			StatusCode: resp.StatusCode,
			Duration:   duration,
			IsSuccess:  false,
			Error:      fmt.Errorf("failed to read response body: %w", err),
		}, err
	}

	// Build response
	httpResp := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Duration:   duration,
		IsSuccess:  resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	if !httpResp.IsSuccess {
		httpResp.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return httpResp, nil
}
