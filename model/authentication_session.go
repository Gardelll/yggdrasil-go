/*
 * Copyright (C) 2022. Gardel <sunxinao@hotmail.com> and contributors
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

package model

import "time"

type AuthenticationSession struct {
	ServerId string
	Token    Token
	Ip       string
	createAt int64
}

func NewAuthenticationSession(serverId string, token *Token, ip string) (session AuthenticationSession) {
	session.ServerId = serverId
	session.Token = *token
	session.Ip = ip
	session.createAt = time.Now().UnixMilli()
	return session
}

func (s *AuthenticationSession) HasExpired() bool {
	d := time.Now().Sub(time.UnixMilli(s.createAt))
	return d > 30*time.Second
}
