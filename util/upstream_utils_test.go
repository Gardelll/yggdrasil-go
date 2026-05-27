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
	"strings"
	"testing"

	"gopkg.in/ini.v1"
)

// TestReplaceURLPlaceholders tests URL placeholder replacement
func TestReplaceURLPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		template string
		params   map[string]string
		expected string
	}{
		{
			name:     "uuid placeholder",
			template: "https://api.com/profile/{uuid}",
			params:   map[string]string{"uuid": "abc123"},
			expected: "https://api.com/profile/abc123",
		},
		{
			name:     "username placeholder",
			template: "https://api.com/user/{username}",
			params:   map[string]string{"username": "Steve"},
			expected: "https://api.com/user/Steve",
		},
		{
			name:     "serverId placeholder",
			template: "https://api.com/server/{serverId}",
			params:   map[string]string{"serverId": "srv001"},
			expected: "https://api.com/server/srv001",
		},
		{
			name:     "ip placeholder",
			template: "https://api.com/check/{ip}",
			params:   map[string]string{"ip": "192.168.1.1"},
			expected: "https://api.com/check/192.168.1.1",
		},
		{
			name:     "multiple placeholders",
			template: "https://api.com/user/{username}/server/{serverId}",
			params:   map[string]string{"username": "Steve", "serverId": "srv001"},
			expected: "https://api.com/user/Steve/server/srv001",
		},
		{
			name:     "with query parameters",
			template: "https://api.com/profile/{uuid}?unsigned=true",
			params:   map[string]string{"uuid": "abc123"},
			expected: "https://api.com/profile/abc123?unsigned=true",
		},
		{
			name:     "url encoding required",
			template: "https://api.com/user/{username}",
			params:   map[string]string{"username": "User Name"},
			expected: "https://api.com/user/User%20Name",
		},
		{
			name:     "no placeholders",
			template: "https://api.com/publickeys",
			params:   map[string]string{},
			expected: "https://api.com/publickeys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceURLPlaceholders(tt.template, tt.params)
			if result != tt.expected {
				t.Errorf("ReplaceURLPlaceholders() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestValidateProxyURL tests proxy URL validation
func TestValidateProxyURL(t *testing.T) {
	tests := []struct {
		name      string
		proxyURL  string
		wantError bool
	}{
		{"http no auth", "http://proxy.example.com:8080", false},
		{"http with auth", "http://user:pass@proxy.example.com:8080", false},
		{"https", "https://proxy.example.com:8443", false},
		{"socks5 no auth", "socks5://127.0.0.1:1080", false},
		{"socks5 with auth", "socks5://user:pass@127.0.0.1:1080", false},
		{"empty string", "", true},
		{"unsupported scheme ftp", "ftp://proxy.example.com:21", true},
		{"unsupported scheme socks4", "socks4://127.0.0.1:1080", true},
		{"missing host", "socks5://", true},
		{"not a url", "not-a-url", true},
		{"http with leading/trailing whitespace", "  http://proxy.example.com:8080  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProxyURL(tt.proxyURL)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateProxyURL(%q) error = %v, wantError %v", tt.proxyURL, err, tt.wantError)
			}
		})
	}
}

// TestResolveUpstreamProxy tests per-upstream proxy resolution against the default.
func TestResolveUpstreamProxy(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		defaultProxy string
		want         string
		wantError    bool
	}{
		{"empty raw, empty default", "", "", "", false},
		{"empty raw inherits default", "", "socks5://127.0.0.1:1080", "socks5://127.0.0.1:1080", false},
		{"direct overrides empty default", "direct", "", "", false},
		{"direct overrides set default", "direct", "socks5://127.0.0.1:1080", "", false},
		{"raw url overrides default", "http://proxy:8080", "socks5://127.0.0.1:1080", "http://proxy:8080", false},
		{"raw url with no default", "socks5://user:pass@host:1080", "", "socks5://user:pass@host:1080", false},
		{"invalid raw url", "ftp://host:21", "", "", true},
		{"whitespace trimmed around direct", "  direct  ", "socks5://127.0.0.1:1080", "", false},
		{"uppercase Direct keyword", "Direct", "socks5://127.0.0.1:1080", "", false},
		{"whitespace-padded default is trimmed", "", "  http://proxy:8080  ", "http://proxy:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveUpstreamProxy(tt.raw, tt.defaultProxy)
			if (err != nil) != tt.wantError {
				t.Fatalf("ResolveUpstreamProxy(%q, %q) error = %v, wantError %v", tt.raw, tt.defaultProxy, err, tt.wantError)
			}
			if got != tt.want {
				t.Errorf("ResolveUpstreamProxy(%q, %q) = %q, want %q", tt.raw, tt.defaultProxy, got, tt.want)
			}
		})
	}
}

// TestValidateURLTemplate tests URL template validation
func TestValidateURLTemplate(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		wantError bool
	}{
		{
			name:      "valid http URL",
			template:  "http://api.com/profile/{uuid}",
			wantError: false,
		},
		{
			name:      "valid https URL",
			template:  "https://api.com/profile/{uuid}",
			wantError: false,
		},
		{
			name:      "invalid protocol",
			template:  "ftp://api.com/profile/{uuid}",
			wantError: true,
		},
		{
			name:      "no protocol",
			template:  "api.com/profile/{uuid}",
			wantError: true,
		},
		{
			name:      "balanced braces",
			template:  "https://api.com/{path}/{uuid}",
			wantError: false,
		},
		{
			name:      "unbalanced braces - missing close",
			template:  "https://api.com/{uuid",
			wantError: true,
		},
		{
			name:      "unbalanced braces - missing open",
			template:  "https://api.com/uuid}",
			wantError: true,
		},
		{
			name:      "no placeholders",
			template:  "https://api.com/publickeys",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURLTemplate(tt.template)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateURLTemplate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestParseUpstreamConfig_Proxy verifies default_proxy / proxy fields end-to-end.
func TestParseUpstreamConfig_Proxy(t *testing.T) {
	const iniSrc = `
[upstream]
services = a, b, c
pool_size = 100
retry_interval = 100
recovery_timeout = 600000
default_proxy = socks5://127.0.0.1:1080

[upstream_a]
profile_url = https://a.example.com/profile/{uuid}
lookup_by_name_url = https://a.example.com/byname/{username}
lookup_by_uuid_url = https://a.example.com/byuuid/{uuid}
bulk_lookup_url = https://a.example.com/bulk
join_url = https://a.example.com/join
has_joined_url = https://a.example.com/hasJoined
public_keys_url = https://a.example.com/publickeys
timeout = 10000

[upstream_b]
profile_url = https://b.example.com/profile/{uuid}
lookup_by_name_url = https://b.example.com/byname/{username}
lookup_by_uuid_url = https://b.example.com/byuuid/{uuid}
bulk_lookup_url = https://b.example.com/bulk
join_url = https://b.example.com/join
has_joined_url = https://b.example.com/hasJoined
public_keys_url = https://b.example.com/publickeys
timeout = 10000
proxy = http://override.example.com:8080

[upstream_c]
profile_url = https://c.example.com/profile/{uuid}
lookup_by_name_url = https://c.example.com/byname/{username}
lookup_by_uuid_url = https://c.example.com/byuuid/{uuid}
bulk_lookup_url = https://c.example.com/bulk
join_url = https://c.example.com/join
has_joined_url = https://c.example.com/hasJoined
public_keys_url = https://c.example.com/publickeys
timeout = 10000
proxy = direct
`
	cfg, err := ini.Load([]byte(iniSrc))
	if err != nil {
		t.Fatalf("ini.Load: %v", err)
	}
	var upstreamCfg UpstreamConfig
	if err := cfg.Section("upstream").MapTo(&upstreamCfg); err != nil {
		t.Fatalf("MapTo: %v", err)
	}

	got, err := ParseUpstreamConfig(&upstreamCfg, cfg)
	if err != nil {
		t.Fatalf("ParseUpstreamConfig: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d upstreams, want 3", len(got))
	}

	byID := map[string]*UpstreamServiceConfig{}
	for _, u := range got {
		byID[u.Id] = u
	}
	if byID["a"].Proxy != "socks5://127.0.0.1:1080" {
		t.Errorf("a inherits default: Proxy = %q, want socks5://127.0.0.1:1080", byID["a"].Proxy)
	}
	if byID["b"].Proxy != "http://override.example.com:8080" {
		t.Errorf("b overrides: Proxy = %q, want http://override.example.com:8080", byID["b"].Proxy)
	}
	if byID["c"].Proxy != "" {
		t.Errorf("c is direct: Proxy = %q, want empty", byID["c"].Proxy)
	}
}

// TestParseUpstreamConfig_InvalidDefaultProxy verifies fail-fast on bad default_proxy.
func TestParseUpstreamConfig_InvalidDefaultProxy(t *testing.T) {
	const iniSrc = `
[upstream]
services = a
default_proxy = ftp://nope:21

[upstream_a]
profile_url = https://a.example.com/profile/{uuid}
lookup_by_name_url = https://a.example.com/byname/{username}
lookup_by_uuid_url = https://a.example.com/byuuid/{uuid}
bulk_lookup_url = https://a.example.com/bulk
join_url = https://a.example.com/join
has_joined_url = https://a.example.com/hasJoined
public_keys_url = https://a.example.com/publickeys
timeout = 10000
`
	cfg, err := ini.Load([]byte(iniSrc))
	if err != nil {
		t.Fatalf("ini.Load: %v", err)
	}
	var upstreamCfg UpstreamConfig
	if err := cfg.Section("upstream").MapTo(&upstreamCfg); err != nil {
		t.Fatalf("MapTo: %v", err)
	}
	if _, err := ParseUpstreamConfig(&upstreamCfg, cfg); err == nil {
		t.Error("expected error for invalid default_proxy, got nil")
	}
}

// TestParseUpstreamConfig_InvalidDefaultProxy_EmptyServices verifies fail-fast on bad
// default_proxy even when no upstream actually inherits it (services list is empty).
func TestParseUpstreamConfig_InvalidDefaultProxy_EmptyServices(t *testing.T) {
	const iniSrc = `
[upstream]
services =
default_proxy = ftp://nope:21
`
	cfg, err := ini.Load([]byte(iniSrc))
	if err != nil {
		t.Fatalf("ini.Load: %v", err)
	}
	var upstreamCfg UpstreamConfig
	if err := cfg.Section("upstream").MapTo(&upstreamCfg); err != nil {
		t.Fatalf("MapTo: %v", err)
	}
	if _, err := ParseUpstreamConfig(&upstreamCfg, cfg); err == nil {
		t.Error("expected error for invalid default_proxy even when services empty, got nil")
	}
}

// TestParseUpstreamConfig_InvalidUpstreamProxy verifies fail-fast on bad per-upstream proxy.
func TestParseUpstreamConfig_InvalidUpstreamProxy(t *testing.T) {
	const iniSrc = `
[upstream]
services = a

[upstream_a]
profile_url = https://a.example.com/profile/{uuid}
lookup_by_name_url = https://a.example.com/byname/{username}
lookup_by_uuid_url = https://a.example.com/byuuid/{uuid}
bulk_lookup_url = https://a.example.com/bulk
join_url = https://a.example.com/join
has_joined_url = https://a.example.com/hasJoined
public_keys_url = https://a.example.com/publickeys
timeout = 10000
proxy = not-a-url
`
	cfg, err := ini.Load([]byte(iniSrc))
	if err != nil {
		t.Fatalf("ini.Load: %v", err)
	}
	var upstreamCfg UpstreamConfig
	if err := cfg.Section("upstream").MapTo(&upstreamCfg); err != nil {
		t.Fatalf("MapTo: %v", err)
	}
	_, err = ParseUpstreamConfig(&upstreamCfg, cfg)
	if err == nil {
		t.Fatal("expected error for invalid per-upstream proxy, got nil")
	}
	if !strings.Contains(err.Error(), "upstream_a") {
		t.Errorf("error should mention section upstream_a, got: %v", err)
	}
}
