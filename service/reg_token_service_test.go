/*
 * Copyright (C) 2025. Gardel <sunxinao@hotmail.com> and contributors
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
	"strings"
	"testing"
)

func TestRemoveHtmlComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "单行注释",
			input:    "<div><!-- This is a comment -->Hello</div>",
			expected: "<div>Hello</div>",
		},
		{
			name: "多行注释",
			input: `<div>
<!-- This is a 
multi-line comment -->
Hello</div>`,
			expected: `<div>

Hello</div>`,
		},
		{
			name:     "多个注释",
			input:    "<!-- comment1 --><p>Text</p><!-- comment2 -->",
			expected: "<p>Text</p>",
		},
		{
			name:     "嵌套标签中的注释",
			input:    "<html><!-- head comment --><body><!-- body comment -->Content</body></html>",
			expected: "<html><body>Content</body></html>",
		},
		{
			name:     "无注释",
			input:    "<div>No comments here</div>",
			expected: "<div>No comments here</div>",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "仅注释",
			input:    "<!-- Only comment -->",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeHtmlComments(tt.input)
			if result != tt.expected {
				t.Errorf("removeHtmlComments() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRemoveHtmlCommentsPreservesTemplateVariables(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<!-- This is a comment -->
<body>
<!-- Another comment -->
<div>Hello {{.Name}}</div>
<!-- End comment -->
</body>
</html>`

	result := removeHtmlComments(input)

	// 应该移除所有注释
	if strings.Contains(result, "<!--") || strings.Contains(result, "-->") {
		t.Error("Comments were not removed")
	}

	// 应该保留模板变量
	if !strings.Contains(result, "{{.Name}}") {
		t.Error("Template variables were not preserved")
	}

	// 应该保留HTML结构
	if !strings.Contains(result, "<body>") || !strings.Contains(result, "</body>") {
		t.Error("HTML structure was not preserved")
	}
}
