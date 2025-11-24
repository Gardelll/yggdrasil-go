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
