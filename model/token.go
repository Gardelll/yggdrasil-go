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

import (
	"time"
	"yggdrasil-go/util"
)

type Token struct {
	createAt        int64
	ClientToken     string
	AccessToken     string
	SelectedProfile Profile
}

type AvailableLevel uint

const (
	Valid AvailableLevel = iota
	NeedRefresh
	Invalid
)

func NewToken(accessToken string, clientToken *string, selectedProfile *Profile) (this Token) {
	this.createAt = time.Now().UnixMilli()

	if clientToken == nil || (len(*clientToken) == 0) {
		this.ClientToken = util.RandomUUID()
	} else {
		this.ClientToken = *clientToken
	}
	this.AccessToken = accessToken
	this.SelectedProfile = *selectedProfile
	return this
}

func (t *Token) GetAvailableLevel() AvailableLevel {
	d := time.Now().Sub(time.UnixMilli(t.createAt))
	if d > time.Hour*24*30 {
		return Invalid
	} else if d > time.Hour*24*15 {
		return NeedRefresh
	} else {
		return Valid
	}
}
