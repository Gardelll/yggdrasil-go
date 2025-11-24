/*
 * Copyright (C) 2025. Gardel <gardel741@outlook.com> and contributors
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

type RegToken struct {
	createAt    int64
	AccessToken string
	Email       string
}

func NewRegToken(email string) (this RegToken) {
	this.Email = email
	this.AccessToken = util.RandomUUID()
	this.createAt = time.Now().UnixMilli()
	return this
}

func (t *RegToken) IsValid() bool {
	d := time.Now().Sub(time.UnixMilli(t.createAt))
	return d < 10*time.Minute
}
