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

package util

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var MessageInvalidToken = "Invalid token."
var MessageInvalidCredentials = "Invalid credentials. Invalid username or password."
var MessageTokenAlreadyAssigned = "Access token already has a profile assigned."
var MessageAccessDenied = "Access denied."
var MessageProfileNotFound = "No such profile."

type YggdrasilError struct {
	ErrorCode    string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
	Cause        string `json:"cause,omitempty"`
	Status       int    `json:"-"`
}

func (e YggdrasilError) Error() string {
	return e.ErrorMessage
}

func NewIllegalArgumentError(msg string) (err YggdrasilError) {
	err.ErrorCode = "IllegalArgumentException"
	err.Status = http.StatusBadRequest
	err.ErrorMessage = msg
	return err
}

func NewForbiddenOperationError(msg string) (err YggdrasilError) {
	err.ErrorCode = "ForbiddenOperationException"
	err.Status = http.StatusForbidden
	err.ErrorMessage = msg
	return err
}

func HandleError(c *gin.Context, err error) {
	switch x := err.(type) {
	case YggdrasilError:
		if x.Status == 0 {
			x.Status = http.StatusForbidden
		}
		if x.Status == http.StatusNoContent {
			c.Status(x.Status)
		} else {
			c.AbortWithStatusJSON(x.Status, x)
		}
		break
	case *YggdrasilError:
		if x.Status == 0 {
			x.Status = http.StatusForbidden
		}
		if x.Status == http.StatusNoContent {
			c.Status(x.Status)
		} else {
			c.AbortWithStatusJSON(x.Status, x)
		}
		break
	default:
		c.AbortWithStatusJSON(http.StatusForbidden, x.Error())
		break
	}
}
