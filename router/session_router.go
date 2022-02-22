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

package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"yggdrasil-go/service"
	"yggdrasil-go/util"
)

type SessionRouter interface {
	JoinServer(c *gin.Context)
	HasJoinedServer(c *gin.Context)
}

type sessionRouterImpl struct {
	sessionService service.SessionService
	skinRootUrl    string
}

func NewSessionRouter(sessionService service.SessionService, skinRootUrl string) SessionRouter {
	sessionRouter := sessionRouterImpl{
		sessionService: sessionService,
		skinRootUrl:    skinRootUrl,
	}
	return &sessionRouter
}

type JoinServerRequest struct {
	AccessToken     string `json:"accessToken" binding:"required"`
	SelectedProfile string `json:"selectedProfile" binding:"required"`
	ServerId        string `json:"serverId" binding:"required"`
}

func (s *sessionRouterImpl) JoinServer(c *gin.Context) {
	request := JoinServerRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	ip := c.Request.RemoteAddr[:strings.LastIndexByte(c.Request.RemoteAddr, ':')]
	err = s.sessionService.JoinServer(request.AccessToken, request.ServerId, request.SelectedProfile, ip)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *sessionRouterImpl) HasJoinedServer(c *gin.Context) {
	username := c.Query("username")
	serverId := c.Query("serverId")
	ip := c.DefaultQuery("ip",
		c.Request.RemoteAddr[:strings.LastIndexByte(c.Request.RemoteAddr, ':')])
	var textureBaseUrl string
	if len(s.skinRootUrl) > 0 {
		textureBaseUrl = strings.TrimRight(s.skinRootUrl, "/") + "/textures"
	} else {
		textureBaseUrl = c.Request.URL.Scheme + "://" + c.Request.URL.Hostname() + "/textures"
	}
	response, err := s.sessionService.HasJoinedServer(serverId, username, ip, textureBaseUrl)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}
