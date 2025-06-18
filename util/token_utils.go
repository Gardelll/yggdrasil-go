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

package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

type officialTokenPayload struct {
	Xuid     string        `json:"xuid"`
	Agg      string        `json:"agg"`
	Sub      string        `json:"sub"`
	Auth     string        `json:"auth"`
	Ns       string        `json:"ns"`
	Roles    []interface{} `json:"roles"`
	Iss      string        `json:"iss"`
	Flags    []string      `json:"flags"`
	Profiles struct {
		Mc string `json:"mc"`
	} `json:"profiles"`
	Platform string `json:"platform"`
	Pfd      []struct {
		Type string `json:"type"`
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"pfd"`
	Nbf int `json:"nbf"`
	Exp int `json:"exp"`
	Iat int `json:"iat"`
}

func ParseOfficialToken(token string) (id, name string, err error) {
	firstDot := strings.IndexRune(token, '.')
	if firstDot == -1 {
		return id, name, errors.New("invalid token")
	}
	secondDot := 1 + firstDot + strings.IndexRune(token[firstDot+1:], '.')
	if secondDot == -1 {
		return id, name, errors.New("invalid token")
	}
	jsonBase64 := token[firstDot+1 : secondDot]
	jsonDecoded, err := base64.RawURLEncoding.DecodeString(jsonBase64)
	if err != nil {
		return id, name, err
	}
	payload := officialTokenPayload{}
	err = json.Unmarshal(jsonDecoded, &payload)
	if err != nil {
		return id, name, err
	}
	if payload.Pfd == nil || len(payload.Pfd) == 0 {
		return id, name, errors.New("invalid token")
	}
	id = payload.Pfd[0].Id
	name = payload.Pfd[0].Name
	return id, name, nil
}
