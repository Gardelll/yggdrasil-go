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
	"yggdrasil-go/model"
	"yggdrasil-go/service"
	"yggdrasil-go/util"
)

type TextureRouter interface {
	GetTexture(c *gin.Context)
	SetTexture(c *gin.Context)
	UploadTexture(c *gin.Context)
	DeleteTexture(c *gin.Context)
}

type textureRouterImpl struct {
	textureService service.TextureService
}

func NewTextureRouter(textureService service.TextureService) TextureRouter {
	textureRouter := textureRouterImpl{textureService: textureService}
	return &textureRouter
}

type SetTextureRequest struct {
	Url   string `json:"url" binding:"required,url"`
	Model string `json:"model" binding:"oneof=slim default"`
}

func (t *textureRouterImpl) GetTexture(c *gin.Context) {
	hash := c.Param("hash")
	response, err := t.textureService.GetTexture(hash)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Data(http.StatusOK, "image/png", response)
}

func (t *textureRouterImpl) SetTexture(c *gin.Context) {
	request := SetTextureRequest{Model: string(model.STEVE)}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, util.NewForbiddenOperationError(err.Error()))
		return
	}
	bearerToken := c.GetHeader("Authorization")
	if len(bearerToken) < 8 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, util.NewForbiddenOperationError(util.MessageInvalidToken))
		return
	}
	accessToken := bearerToken[7:]
	profileId, err := util.ToUUID(c.Param("uuid"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError(err.Error()))
		return
	}
	textureType := c.Param("textureType")
	if textureType != "skin" && textureType != "cape" {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError("Invalid texture type."))
		return
	}
	if request.Model == "" {
		request.Model = string(model.STEVE)
	}
	modelType := model.ModelType(request.Model)
	err = t.textureService.SetTexture(accessToken, profileId, request.Url, textureType, &modelType)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (t *textureRouterImpl) UploadTexture(c *gin.Context) {
	bearerToken := c.GetHeader("Authorization")
	if len(bearerToken) < 8 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, util.NewForbiddenOperationError(util.MessageInvalidToken))
		return
	}
	accessToken := bearerToken[7:]
	profileId, err := util.ToUUID(c.Param("uuid"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError(err.Error()))
		return
	}
	textureType := c.Param("textureType")
	if textureType != "skin" && textureType != "cape" {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError("Invalid texture type."))
		return
	}
	modelStr := c.PostForm("model")
	modelType := model.STEVE
	if modelStr == "ALEX" {
		modelType = model.ALEX
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError(err.Error()))
		return
	}
	if file.Size > (1 << 20) {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError("File too large(more than 1MiB)"))
		return
	}
	fileReader, err := file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, util.YggdrasilError{
			ErrorCode:    "Internal Server Error",
			ErrorMessage: "Can not open file.",
		})
		return
	}
	defer fileReader.Close()
	err = t.textureService.UploadTexture(accessToken, profileId, fileReader, textureType, &modelType)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (t *textureRouterImpl) DeleteTexture(c *gin.Context) {
	bearerToken := c.GetHeader("Authorization")
	if len(bearerToken) < 8 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, util.NewForbiddenOperationError(util.MessageInvalidToken))
		return
	}
	accessToken := bearerToken[7:]
	profileId, err := util.ToUUID(c.Param("uuid"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError(err.Error()))
		return
	}
	textureType := c.Param("textureType")
	if textureType != "skin" && textureType != "cape" {
		c.AbortWithStatusJSON(http.StatusBadRequest, util.NewIllegalArgumentError("Invalid texture type."))
		return
	}
	err = t.textureService.DeleteTexture(accessToken, profileId, textureType)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
