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
	"net/http"
	"testing"

	"yggdrasil-go/util"
)

// TestNewUpstreamService_PerUpstreamClient verifies that each upstream gets its own
// *http.Client and that proxy configuration is honored.
func TestNewUpstreamService_PerUpstreamClient(t *testing.T) {
	cfg := &util.UpstreamConfig{
		Services:        []string{"a", "b"},
		PoolSize:        10,
		RetryInterval:   100,
		RecoveryTimeout: 60000,
	}
	upstreams := []*util.UpstreamServiceConfig{
		{
			Id:              "a",
			ProfileURL:      "https://a/profile/{uuid}",
			LookupByNameURL: "https://a/byname/{username}",
			LookupByUUIDURL: "https://a/byuuid/{uuid}",
			BulkLookupURL:   "https://a/bulk",
			JoinURL:         "https://a/join",
			HasJoinedURL:    "https://a/hasJoined",
			PublicKeysURL:   "https://a/publickeys",
			Timeout:         10000,
			Proxy:           "",
		},
		{
			Id:              "b",
			ProfileURL:      "https://b/profile/{uuid}",
			LookupByNameURL: "https://b/byname/{username}",
			LookupByUUIDURL: "https://b/byuuid/{uuid}",
			BulkLookupURL:   "https://b/bulk",
			JoinURL:         "https://b/join",
			HasJoinedURL:    "https://b/hasJoined",
			PublicKeysURL:   "https://b/publickeys",
			Timeout:         10000,
			Proxy:           "socks5://127.0.0.1:1080",
		},
	}

	svc, err := NewUpstreamService(cfg, upstreams)
	if err != nil {
		t.Fatalf("NewUpstreamService: %v", err)
	}

	impl, ok := svc.(*upstreamService)
	if !ok {
		t.Fatalf("svc is not *upstreamService, got %T", svc)
	}
	stateA := impl.upstreams["a"]
	stateB := impl.upstreams["b"]
	if stateA == nil || stateB == nil {
		t.Fatal("expected both upstreams to be present")
	}
	if stateA.client == nil || stateB.client == nil {
		t.Fatal("each upstream must have its own client")
	}
	if stateA.client == stateB.client {
		t.Error("upstreams should not share the same *http.Client")
	}
	if stateA.Proxy != "" {
		t.Errorf("a.Proxy = %q, want empty", stateA.Proxy)
	}
	if stateB.Proxy != "socks5://127.0.0.1:1080" {
		t.Errorf("b.Proxy = %q, want socks5://127.0.0.1:1080", stateB.Proxy)
	}

	// Verify b's client transport actually has SOCKS5 wiring (DialContext non-nil)
	// — not just that Proxy field was copied.
	trB, ok := stateB.client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("stateB.client.Transport is not *http.Transport, got %T", stateB.client.Transport)
	}
	if trB.DialContext == nil {
		t.Error("stateB.client.Transport.DialContext should be set for socks5 proxy")
	}
}

// TestNewUpstreamService_InvalidProxy verifies fail-fast on an invalid proxy URL.
func TestNewUpstreamService_InvalidProxy(t *testing.T) {
	cfg := &util.UpstreamConfig{
		Services:        []string{"a"},
		PoolSize:        10,
		RetryInterval:   100,
		RecoveryTimeout: 60000,
	}
	upstreams := []*util.UpstreamServiceConfig{
		{
			Id:              "a",
			ProfileURL:      "https://a/profile/{uuid}",
			LookupByNameURL: "https://a/byname/{username}",
			LookupByUUIDURL: "https://a/byuuid/{uuid}",
			BulkLookupURL:   "https://a/bulk",
			JoinURL:         "https://a/join",
			HasJoinedURL:    "https://a/hasJoined",
			PublicKeysURL:   "https://a/publickeys",
			Timeout:         10000,
			Proxy:           "ftp://nope:21",
		},
	}
	if _, err := NewUpstreamService(cfg, upstreams); err == nil {
		t.Error("expected error for invalid proxy URL, got nil")
	}
}

// TestRedactProxyForLog verifies userinfo is stripped from proxy URLs in log output.
func TestRedactProxyForLog(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty is direct", "", "direct"},
		{"no userinfo passes through", "socks5://proxy.example.com:1080", "socks5://proxy.example.com:1080"},
		{"userinfo stripped from socks5", "socks5://user:secret@proxy.example.com:1080", "socks5://proxy.example.com:1080"},
		{"userinfo stripped from http", "http://admin:pw@proxy.corp:8080", "http://proxy.corp:8080"},
		{"username only stripped", "http://user@proxy:8080", "http://proxy:8080"},
		{"invalid url returns sentinel", "://not-a-url", "<invalid>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactProxyForLog(tt.input)
			if got != tt.want {
				t.Errorf("redactProxyForLog(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
