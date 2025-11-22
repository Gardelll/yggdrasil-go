/*
 * Copyright (C) 2022-2025. Gardel <sunxinao@hotmail.com> and contributors
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
	"bytes"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"yggdrasil-go/model"
	"yggdrasil-go/util"
)

type TextureService interface {
	GetTexture(hash string) ([]byte, error)
	SetTexture(accessToken string, profileId uuid.UUID, skinUrl string, textureType string, model *model.ModelType) error
	UploadTexture(accessToken string, profileId uuid.UUID, skinReader io.Reader, textureType string, model *model.ModelType) error
	DeleteTexture(accessToken string, profileId uuid.UUID, textureType string) error
}

type textureServiceImpl struct {
	tokenService TokenService
	db           *gorm.DB
}

func NewTextureService(tokenService TokenService, db *gorm.DB) TextureService {
	textureService := textureServiceImpl{
		tokenService: tokenService,
		db:           db,
	}
	return &textureService
}

func (t *textureServiceImpl) GetTexture(hash string) ([]byte, error) {
	texture := model.Texture{}
	if err := t.db.First(&texture, "hash = ?", hash).Error; err == nil {
		return texture.Data, nil
	} else {
		err := util.YggdrasilError{
			Status:       http.StatusNotFound,
			ErrorCode:    "Not Found",
			ErrorMessage: "Texture Not Found",
		}
		return nil, &err
	}
}

func (t *textureServiceImpl) SetTexture(accessToken string, profileId uuid.UUID, skinUrl string, textureType string, modelType *model.ModelType) error {
	token, ok := t.tokenService.GetToken(accessToken)
	if !ok || token.GetAvailableLevel() != model.Valid {
		return util.NewForbiddenOperationError(util.MessageInvalidToken)
	}
	if token.SelectedProfile.Id != profileId {
		return util.NewForbiddenOperationError("Profile mismatch.")
	}
	user := model.User{}
	if err := t.db.First(&user, profileId).Error; err != nil {
		return util.NewForbiddenOperationError(util.MessageProfileNotFound)
	}
	skinDownloadUrl, err := url.Parse(skinUrl)
	if err != nil {
		return util.NewIllegalArgumentError("Invalid skin url: " + err.Error())
	}
	response, err := http.Get(skinDownloadUrl.String())
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.ContentLength > 1048576 {
		return util.NewIllegalArgumentError("File too large(more than 1MiB)")
	}
	reader := io.LimitReader(response.Body, 1048576)
	var header bytes.Buffer
	conf, _, err := image.DecodeConfig(io.TeeReader(reader, &header))
	if err != nil || conf.Width > 1024 || conf.Height > 1024 {
		return util.NewIllegalArgumentError("Image too large(max 1024 pixels each dimension)")
	}
	im, _, err := image.Decode(io.MultiReader(&header, reader))
	if err != nil {
		return err
	}
	err = t.saveTexture(&user, im, textureType, modelType)
	if err != nil {
		return err
	} else {
		profile, _ := user.Profile()
		token.SelectedProfile = *profile
		return nil
	}
}

func (t *textureServiceImpl) UploadTexture(accessToken string, profileId uuid.UUID, skinReader io.Reader, textureType string, modelType *model.ModelType) error {
	token, ok := t.tokenService.GetToken(accessToken)
	if !ok || token.GetAvailableLevel() != model.Valid {
		return util.NewForbiddenOperationError(util.MessageInvalidToken)
	}
	if token.SelectedProfile.Id != profileId {
		return util.NewForbiddenOperationError("Profile mismatch.")
	}
	user := model.User{}
	if err := t.db.First(&user, profileId).Error; err != nil {
		return util.NewForbiddenOperationError(util.MessageProfileNotFound)
	}
	reader := io.LimitReader(skinReader, 1048576)
	var header bytes.Buffer
	conf, _, err := image.DecodeConfig(io.TeeReader(reader, &header))
	if err != nil || conf.Width > 1024 || conf.Height > 1024 {
		return util.NewIllegalArgumentError("Image too large(max 1024 pixels each dimension)")
	}
	im, _, err := image.Decode(io.MultiReader(&header, reader))
	if err != nil {
		return err
	}
	err = t.saveTexture(&user, im, textureType, modelType)
	if err != nil {
		return err
	} else {
		profile, _ := user.Profile()
		token.SelectedProfile = *profile
		return nil
	}
}

func (t *textureServiceImpl) DeleteTexture(accessToken string, profileId uuid.UUID, textureType string) error {
	token, ok := t.tokenService.GetToken(accessToken)
	if !ok || token.GetAvailableLevel() != model.Valid {
		return util.NewForbiddenOperationError(util.MessageInvalidToken)
	}
	if token.SelectedProfile.Id != profileId {
		return util.NewForbiddenOperationError("Profile mismatch.")
	}
	user := model.User{}
	if err := t.db.First(&user, profileId).Error; err != nil {
		return util.NewForbiddenOperationError(util.MessageProfileNotFound)
	}
	textureType = strings.ToUpper(textureType)
	if textureType != "SKIN" && textureType != "CAPE" {
		textureType = "SKIN"
	}
	var profile *model.Profile
	hash, ok := token.SelectedProfile.Textures[textureType]
	if ok {
		delete(token.SelectedProfile.Textures, textureType)
		profile = &token.SelectedProfile
	} else {
		p, err := user.Profile()
		if err != nil {
			return err
		}
		hash, ok = p.Textures[textureType]
		if ok {
			delete(p.Textures, textureType)
		} else {
			return util.NewForbiddenOperationError(util.MessageProfileNotFound)
		}
		profile = p
	}
	return t.db.Transaction(func(tx *gorm.DB) error {
		texture := model.Texture{}
		if err := tx.Select("hash", "used").First(&texture, "hash = ?", hash).Error; err == nil {
			if texture.Used < 2 {
				tx.Delete(&texture)
			} else {
				tx.Model(&texture).Update("used", gorm.Expr("used - ?", 1))
			}
		}
		user.SetProfile(profile)
		return tx.Save(&user).Error
	})
}

func (t *textureServiceImpl) saveTexture(user *model.User, skinImage image.Image, textureType string, modelType *model.ModelType) error {
	var modelValue model.ModelType
	if modelType != nil && *modelType == model.ALEX {
		modelValue = *modelType
	} else {
		modelValue = model.STEVE
	}
	textureType = strings.ToUpper(textureType)
	if textureType != "SKIN" && textureType != "CAPE" {
		textureType = "SKIN"
	}
	return t.db.Transaction(func(tx *gorm.DB) error {
		profile, err := user.Profile()
		if err != nil {
			return err
		}
		if textureType == "SKIN" {
			profile.ModelType = modelValue
		}
		hash := model.ComputeTextureId(skinImage)
		oldHash, oldExist := profile.Textures[textureType]
		texture := model.Texture{}
		if err := tx.First(&texture, "hash = ?", hash).Error; err != nil {
			// Texture does not exist, create new texture with used = 1
			texture.Hash = hash
			texture.Used = 1
			buffer := bytes.Buffer{}
			err := png.Encode(&buffer, skinImage)
			if err != nil {
				return err
			}
			texture.Data = buffer.Bytes()
			if err := tx.Create(&texture).Error; err != nil {
				return err
			}
		} else {
			// Texture already exists
			// Increment reference count if:
			// - First time setting texture (!oldExist), OR
			// - Replacing with different texture (oldHash != hash)
			// Do NOT increment if re-uploading same texture (oldHash == hash)
			if !oldExist || oldHash != hash {
				tx.Model(&texture).Update("used", gorm.Expr("used + ?", 1))
			}
		}
		if oldExist && oldHash != hash {
			oldTexture := model.Texture{}
			if err := tx.Select("hash", "used").First(&oldTexture, "hash = ?", oldHash).Error; err == nil {
				if oldTexture.Used < 2 {
					tx.Delete(&oldTexture)
				} else {
					tx.Model(&oldTexture).Update("used", gorm.Expr("used - ?", 1))
				}
			}
		}
		profile.Textures[textureType] = hash
		user.SetProfile(profile)
		return tx.Save(&user).Error
	})
}
