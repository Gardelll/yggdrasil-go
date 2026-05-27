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

package util

import (
	"errors"
	"fmt"
	"gopkg.in/ini.v1"
	"net/url"
	"strings"
)

// Default Mojang upstream configuration
const (
	DefaultMojangProfileURL      = "https://sessionserver.mojang.com/session/minecraft/profile/{uuid}"
	DefaultMojangLookupByNameURL = "https://api.minecraftservices.com/minecraft/profile/lookup/name/{username}"
	DefaultMojangLookupByUUIDURL = "https://api.minecraftservices.com/minecraft/profile/lookup/{uuid}"
	DefaultMojangBulkLookupURL   = "https://api.minecraftservices.com/minecraft/profile/lookup/bulk/byname"
	DefaultMojangJoinURL         = "https://sessionserver.mojang.com/session/minecraft/join"
	DefaultMojangHasJoinedURL    = "https://sessionserver.mojang.com/session/minecraft/hasJoined"
	DefaultMojangPublicKeysURL   = "https://api.minecraftservices.com/publickeys"
	DefaultMojangTimeout         = 10000
)

// UpstreamConfig represents the upstream configuration from INI file
// Note: ini package supports []string type with comma-separated values
type UpstreamConfig struct {
	Services        []string `ini:"services,omitempty"`
	PoolSize        int      `ini:"pool_size"`
	RetryInterval   int      `ini:"retry_interval"`
	RecoveryTimeout int      `ini:"recovery_timeout"`
	DefaultProxy    string   `ini:"default_proxy,omitempty"`
}

// UpstreamServiceConfig represents a single upstream service configuration
type UpstreamServiceConfig struct {
	Id              string `ini:"-"`                  // Unique identifier (e.g., "mojang")
	ProfileURL      string `ini:"profile_url"`        // Session profile query endpoint (supports {uuid} placeholder)
	LookupByNameURL string `ini:"lookup_by_name_url"` // Lookup by username endpoint (supports {username} placeholder)
	LookupByUUIDURL string `ini:"lookup_by_uuid_url"` // Lookup by UUID endpoint (supports {uuid} placeholder)
	BulkLookupURL   string `ini:"bulk_lookup_url"`    // Bulk lookup endpoint (POST)
	JoinURL         string `ini:"join_url"`           // Join server endpoint (POST)
	HasJoinedURL    string `ini:"has_joined_url"`     // Verify has joined endpoint (supports query parameters)
	PublicKeysURL   string `ini:"public_keys_url"`    // Public keys endpoint
	Timeout         int    `ini:"timeout"`            // Request timeout in milliseconds
	// Proxy holds the effective proxy URL after merging with [upstream].default_proxy.
	// "" means direct (no proxy). Populated by ParseUpstreamConfig.
	Proxy string `ini:"-"`
}

// ParseUpstreamConfig parses the upstream configuration from an INI file
// Behavior:
// - No [upstream] section: Returns default Mojang upstream (backward compatibility)
// - [upstream] section exists but services empty: Returns no upstream (pure local mode)
// - [upstream] section exists with services: Validates and returns configured upstreams
func ParseUpstreamConfig(upstreamCfg *UpstreamConfig, cfg *ini.File) ([]*UpstreamServiceConfig, error) {
	if cfg == nil {
		return nil, errors.New("config file is nil")
	}

	// Parse the main [upstream] section
	hasUpstream := cfg.HasSection("upstream")
	if !hasUpstream {
		// [upstream] section does not exist - return default Mojang upstream for backward compatibility
		mojangConfig := &UpstreamServiceConfig{
			Id:              "mojang",
			ProfileURL:      DefaultMojangProfileURL,
			LookupByNameURL: DefaultMojangLookupByNameURL,
			LookupByUUIDURL: DefaultMojangLookupByUUIDURL,
			BulkLookupURL:   DefaultMojangBulkLookupURL,
			JoinURL:         DefaultMojangJoinURL,
			HasJoinedURL:    DefaultMojangHasJoinedURL,
			PublicKeysURL:   DefaultMojangPublicKeysURL,
			Timeout:         DefaultMojangTimeout,
		}
		return []*UpstreamServiceConfig{mojangConfig}, nil
	}

	// Validate default proxy upfront (fail-fast even if no upstream actually inherits it).
	if upstreamCfg.DefaultProxy != "" {
		if err := ValidateProxyURL(upstreamCfg.DefaultProxy); err != nil {
			return nil, fmt.Errorf("section 'upstream' has invalid default_proxy: %w", err)
		}
	}

	// If [upstream] section exists but services is empty, return pure local mode (no upstreams)
	if len(upstreamCfg.Services) == 0 {
		return nil, nil
	}

	// Parse individual upstream service configurations
	upstreamConfigs := make([]*UpstreamServiceConfig, 0, len(upstreamCfg.Services))
	for _, serviceID := range upstreamCfg.Services {
		sectionName := fmt.Sprintf("upstream_%s", serviceID)
		section, err := cfg.GetSection(sectionName)
		if err != nil {
			return nil, fmt.Errorf("upstream service '%s' is listed in services but section [%s] is missing", serviceID, sectionName)
		}

		serviceConfig := &UpstreamServiceConfig{
			Id: serviceID,
		}
		err = section.MapTo(serviceConfig)
		if err != nil {
			return nil, fmt.Errorf("upstream service '%s' is listed in services but section [%s] is invalid: %w", serviceID, sectionName, err)
		}

		if err = ValidateURLTemplate(serviceConfig.ProfileURL); err != nil {
			return nil, fmt.Errorf("section '%s' has invalid %s: %w", sectionName, "profile_url", err)
		}
		if err = ValidateURLTemplate(serviceConfig.LookupByNameURL); err != nil {
			return nil, fmt.Errorf("section '%s' has invalid %s: %w", sectionName, "lookup_by_name_url", err)
		}
		if err = ValidateURLTemplate(serviceConfig.LookupByUUIDURL); err != nil {
			return nil, fmt.Errorf("section '%s' has invalid %s: %w", sectionName, "lookup_by_uuid_url", err)
		}
		if err = ValidateURLTemplate(serviceConfig.BulkLookupURL); err != nil {
			return nil, fmt.Errorf("section '%s' has invalid %s: %w", sectionName, "bulk_lookup_url", err)
		}
		if err = ValidateURLTemplate(serviceConfig.JoinURL); err != nil {
			return nil, fmt.Errorf("section '%s' has invalid %s: %w", sectionName, "join_url", err)
		}
		if err = ValidateURLTemplate(serviceConfig.HasJoinedURL); err != nil {
			return nil, fmt.Errorf("section '%s' has invalid %s: %w", sectionName, "has_joined_url", err)
		}
		if err = ValidateURLTemplate(serviceConfig.PublicKeysURL); err != nil {
			return nil, fmt.Errorf("section '%s' has invalid %s: %w", sectionName, "public_keys_url", err)
		}
		if serviceConfig.Timeout < 1 {
			return nil, fmt.Errorf("section '%s' has invalid %s: %s", sectionName, "timeout", "Must be greater than 0")
		}

		rawProxy := section.Key("proxy").String()
		effectiveProxy, err := ResolveUpstreamProxy(rawProxy, upstreamCfg.DefaultProxy)
		if err != nil {
			return nil, fmt.Errorf("section '%s' has invalid %s: %w", sectionName, "proxy", err)
		}
		serviceConfig.Proxy = effectiveProxy

		upstreamConfigs = append(upstreamConfigs, serviceConfig)
	}

	return upstreamConfigs, nil
}

// ReplaceURLPlaceholders replaces placeholders in URL template with actual values
// Supported placeholders: {uuid}, {username}, {serverId}, {ip}
// Example: ReplaceURLPlaceholders("https://api.com/user/{uuid}", {"uuid": "123"})
//
//	returns: "https://api.com/user/123"
func ReplaceURLPlaceholders(template string, params map[string]string) string {
	result := template
	for key, value := range params {
		placeholder := "{" + key + "}"
		// URL encode the value for path components
		encodedValue := url.PathEscape(value)
		result = strings.Replace(result, placeholder, encodedValue, -1)
	}
	return result
}

// supportedProxySchemes lists scheme values accepted by ValidateProxyURL.
var supportedProxySchemes = map[string]struct{}{
	"http":   {},
	"https":  {},
	"socks5": {},
}

// ValidateProxyURL validates a proxy URL string.
// Accepts http://, https://, socks5:// with an optional userinfo and a non-empty host.
// Returns an error for empty input — callers must handle the "no proxy" case before calling.
func ValidateProxyURL(proxyURL string) error {
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		return errors.New("proxy URL is empty")
	}
	u, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("parse proxy URL: %w", err)
	}
	if _, ok := supportedProxySchemes[u.Scheme]; !ok {
		return fmt.Errorf("unsupported proxy scheme %q (supported: http, https, socks5)", u.Scheme)
	}
	if u.Host == "" {
		return errors.New("proxy URL missing host")
	}
	return nil
}

// ProxyDirectKeyword is the literal value in INI that means "no proxy, even if default is set".
const ProxyDirectKeyword = "direct"

// ResolveUpstreamProxy merges a per-upstream proxy value with the global default.
// Returns the effective proxy URL ("" means direct).
//
// raw == ""                  -> inherits defaultProxy (may itself be "")
// raw == ProxyDirectKeyword  -> "" (forces direct, overrides default)
// raw == "<url>"             -> validated and returned as-is
//
// defaultProxy is assumed to have been validated earlier (or be ""). It is trimmed
// before being returned so the resolved value is always whitespace-normalized.
func ResolveUpstreamProxy(raw, defaultProxy string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return strings.TrimSpace(defaultProxy), nil
	}
	if strings.EqualFold(raw, ProxyDirectKeyword) {
		return "", nil
	}
	if err := ValidateProxyURL(raw); err != nil {
		return "", err
	}
	return raw, nil
}

// ValidateURLTemplate validates URL template format
func ValidateURLTemplate(template string) error {
	if !strings.HasPrefix(template, "http://") && !strings.HasPrefix(template, "https://") {
		return errors.New("URL must start with http:// or https://")
	}

	// Check that braces are balanced
	openCount := strings.Count(template, "{")
	closeCount := strings.Count(template, "}")
	if openCount != closeCount {
		return errors.New("unbalanced placeholders in URL template")
	}

	return nil
}
