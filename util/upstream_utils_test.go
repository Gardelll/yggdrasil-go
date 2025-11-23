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

import "testing"

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
