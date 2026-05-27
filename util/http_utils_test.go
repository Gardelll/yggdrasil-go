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
	"net/http"
	"testing"
	"time"
)

// TestBuildUpstreamHTTPClient_Direct verifies a direct (no-proxy) client.
func TestBuildUpstreamHTTPClient_Direct(t *testing.T) {
	c, err := BuildUpstreamHTTPClient("", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("client is nil")
	}
	if c.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", c.Timeout)
	}
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport is not *http.Transport, got %T", c.Transport)
	}
	if tr.Proxy != nil {
		t.Error("Transport.Proxy should be nil for direct client")
	}
}

// TestBuildUpstreamHTTPClient_HTTPProxy verifies an http:// proxy is wired via Proxy func.
func TestBuildUpstreamHTTPClient_HTTPProxy(t *testing.T) {
	const proxyURL = "http://proxy.example.com:8080"
	c, err := BuildUpstreamHTTPClient(proxyURL, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport is not *http.Transport, got %T", c.Transport)
	}
	if tr.Proxy == nil {
		t.Fatal("Transport.Proxy should be set for http proxy")
	}
	req, _ := http.NewRequest("GET", "http://target.example.com/path", nil)
	got, err := tr.Proxy(req)
	if err != nil {
		t.Fatalf("Proxy func returned error: %v", err)
	}
	if got == nil || got.String() != proxyURL {
		t.Errorf("Proxy func returned %v, want %s", got, proxyURL)
	}
}

// TestBuildUpstreamHTTPClient_SOCKS5 verifies a socks5:// proxy wires DialContext.
func TestBuildUpstreamHTTPClient_SOCKS5(t *testing.T) {
	c, err := BuildUpstreamHTTPClient("socks5://127.0.0.1:1080", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport is not *http.Transport, got %T", c.Transport)
	}
	if tr.Proxy != nil {
		t.Error("Transport.Proxy should be nil when SOCKS5 is used (DialContext handles it)")
	}
	if tr.DialContext == nil {
		t.Error("Transport.DialContext should be set for socks5 proxy")
	}
}

// TestBuildUpstreamHTTPClient_InvalidScheme verifies an unsupported scheme returns an error.
func TestBuildUpstreamHTTPClient_InvalidScheme(t *testing.T) {
	_, err := BuildUpstreamHTTPClient("ftp://nope:21", 5*time.Second)
	if err == nil {
		t.Error("expected error for unsupported scheme, got nil")
	}
}
