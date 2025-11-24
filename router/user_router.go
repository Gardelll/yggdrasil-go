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

package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"yggdrasil-go/dto"
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
	UUIDToUUID(c *gin.Context)
	QueryUUIDs(c *gin.Context)
	QueryProfile(c *gin.Context)
	PlayerAttributes(c *gin.Context)
	PlayerBlockList(c *gin.Context)
	ProfileKey(c *gin.Context)
	SendEmail(c *gin.Context)
	VerifyEmail(c *gin.Context)
	ResetPassword(c *gin.Context)
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

func (u *userRouterImpl) Register(c *gin.Context) {
	request := dto.RegRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	response, err := u.userService.Register(c.Request.Context(), request.Username, request.Password, request.ProfileName, c.ClientIP())
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (u *userRouterImpl) Login(c *gin.Context) {
	request := dto.LoginRequest{}
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
	request := dto.ChangeProfileRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	err = u.userService.ChangeProfile(c.Request.Context(), request.AccessToken, request.ClientToken, request.ChangeTo)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (u *userRouterImpl) Refresh(c *gin.Context) {
	request := dto.RefreshRequest{}
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
	request := dto.ValidateRequest{}
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
	request := dto.InvalidateRequest{}
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
	request := dto.SignoutRequest{}
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
	response, err := u.userService.UsernameToUUID(c.Request.Context(), username)
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

func (u *userRouterImpl) UUIDToUUID(c *gin.Context) {
	profileIdStr := c.Param("uuid")
	profileId, err := util.ToUUID(profileIdStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError(err.Error()))
		return
	}
	response, err := u.userService.UUIDToUUID(c.Request.Context(), profileId)
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
	response, err := u.userService.QueryUUIDs(c.Request.Context(), request)
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
	response, err := u.userService.QueryProfile(c.Request.Context(), profileId, unsigned, textureBaseUrl)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (u *userRouterImpl) PlayerAttributes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"privileges": gin.H{
			"onlineChat": gin.H{
				"enabled": true,
			},
			"multiplayerServer": gin.H{
				"enabled": true,
			},
			"multiplayerRealms": gin.H{
				"enabled": false,
			},
			"telemetry": gin.H{
				"enabled": false,
			},
			"optionalTelemetry": gin.H{
				"enabled": false,
			},
		},
		"profanityFilterPreferences": gin.H{
			"profanityFilterOn": false,
		},
	})
}

func (u *userRouterImpl) PlayerBlockList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"blockedProfiles": []string{},
	})
}

func (u *userRouterImpl) ProfileKey(c *gin.Context) {
	bearerToken := c.GetHeader("Authorization")
	if len(bearerToken) < 8 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, util.NewForbiddenOperationError(util.MessageInvalidToken))
		return
	}
	accessToken := bearerToken[7:]
	response, err := u.userService.ProfileKey(accessToken)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (u *userRouterImpl) SendEmail(c *gin.Context) {
	var request dto.SendEmailRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	var tokenType service.RegTokenType
	switch request.EmailType {
	case "register":
		tokenType = service.RegisterToken
	case "resetPassword":
		tokenType = service.ResetPasswordToken
	}
	err = u.userService.SendEmail(request.Email, tokenType, c.ClientIP())
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (u *userRouterImpl) VerifyEmail(c *gin.Context) {
	token, ok := c.GetQuery("access_token")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError("access_token is required"))
		return
	}
	err := u.userService.VerifyEmail(token)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (u *userRouterImpl) ResetPassword(c *gin.Context) {
	var request dto.PasswordResetRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	err = u.userService.ResetPassword(request.Email, request.Password, request.AccessToken)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
