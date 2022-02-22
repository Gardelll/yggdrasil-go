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
	"yggdrasil-go/model"
	"yggdrasil-go/service"
	"yggdrasil-go/util"
)

type UserRouter interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	ChangeProfile(c *gin.Context)
	Refresh(c *gin.Context)
	Validate(c *gin.Context)
	Invalidate(c *gin.Context)
	Signout(c *gin.Context)
	UsernameToUUID(c *gin.Context)
	QueryUUIDs(c *gin.Context)
	QueryProfile(c *gin.Context)
}

type userRouterImpl struct {
	userService service.UserService
	skinRootUrl string
}

func NewUserRouter(userService service.UserService, skinRootUrl string) UserRouter {
	userRouter := userRouterImpl{
		userService: userService,
		skinRootUrl: skinRootUrl,
	}
	return &userRouter
}

type RegRequest struct {
	Username    string `json:"username" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	ProfileName string `json:"profileName" binding:"required"`
}

type MinecraftAgent struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type ClientTokenBase struct {
	ClientToken *string `json:"clientToken,omitempty"`
}

type AccessTokenBase struct {
	AccessToken string `json:"accessToken" binding:"required"`
}

type DualTokenBase struct {
	AccessTokenBase
	ClientTokenBase
}

type LoginRequest struct {
	ClientTokenBase
	Username    string          `json:"username" binding:"required,email"`
	Password    string          `json:"password" binding:"required"`
	RequestUser bool            `json:"requestUser"`
	Agent       *MinecraftAgent `json:"agent,omitempty"`
}

type RefreshRequest struct {
	DualTokenBase
	RequestUser     bool                   `json:"requestUser"`
	SelectedProfile *model.ProfileResponse `json:"selectedProfile,omitempty"`
}

type ValidateRequest struct {
	DualTokenBase
}

type ChangeProfileRequest struct {
	DualTokenBase
	ChangeTo string `json:"changeTo" binding:"required"`
}

type InvalidateRequest struct {
	AccessTokenBase
}

type SignoutRequest struct {
	Username string `json:"username" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (u *userRouterImpl) Register(c *gin.Context) {
	request := RegRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	response, err := u.userService.Register(request.Username, request.Password, request.ProfileName)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (u *userRouterImpl) Login(c *gin.Context) {
	request := LoginRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	response, err := u.userService.Login(request.Username, request.Password, request.ClientToken, request.RequestUser)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (u *userRouterImpl) ChangeProfile(c *gin.Context) {
	request := ChangeProfileRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	err = u.userService.ChangeProfile(request.AccessToken, request.ClientToken, request.ChangeTo)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (u *userRouterImpl) Refresh(c *gin.Context) {
	request := RefreshRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	response, err := u.userService.Refresh(request.AccessToken, request.ClientToken, request.RequestUser, request.SelectedProfile)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (u *userRouterImpl) Validate(c *gin.Context) {
	request := ValidateRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	err = u.userService.Validate(request.AccessToken, request.ClientToken)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (u *userRouterImpl) Invalidate(c *gin.Context) {
	request := InvalidateRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	err = u.userService.Invalidate(request.AccessToken)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (u *userRouterImpl) Signout(c *gin.Context) {
	request := SignoutRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	err = u.userService.Signout(request.Username, request.Password)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (u *userRouterImpl) UsernameToUUID(c *gin.Context) {
	username := c.Param("username")
	response, err := u.userService.UsernameToUUID(username)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	if response != nil {
		c.JSON(http.StatusOK, response)
	} else {
		c.Status(http.StatusNoContent)
	}
}

func (u *userRouterImpl) QueryUUIDs(c *gin.Context) {
	var request []string
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	response, err := u.userService.QueryUUIDs(request)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (u *userRouterImpl) QueryProfile(c *gin.Context) {
	profileIdStr := c.Param("uuid")
	profileId, err := util.ToUUID(profileIdStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError(err.Error()))
		return
	}
	unsigned := "true" == c.DefaultQuery("unsigned", "false")
	var textureBaseUrl string
	if len(u.skinRootUrl) > 0 {
		textureBaseUrl = strings.TrimRight(u.skinRootUrl, "/") + "/textures"
	} else {
		textureBaseUrl = c.Request.URL.Scheme + "://" + c.Request.URL.Hostname() + "/textures"
	}
	response, err := u.userService.QueryProfile(profileId, unsigned, textureBaseUrl)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}
