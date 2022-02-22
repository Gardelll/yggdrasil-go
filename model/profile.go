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
	"encoding/json"
	"github.com/google/uuid"
	"time"
	"yggdrasil-go/util"
)

type Profile struct {
	Id        uuid.UUID
	Name      string
	ModelType ModelType
	Textures  map[string]string
}

type ModelType string

const (
	STEVE ModelType = "default"
	ALEX            = "slim"
)

type MetadataType map[string]interface{}

type SkinTexture struct {
	Url      string        `json:"url"`
	Metadata *MetadataType `json:"metadata,omitempty"`
}

type CapeTexture struct {
	Url string `json:"url"`
}

type TexturesType struct {
	SKIN *SkinTexture `json:"SKIN,omitempty"`
	CAPE *CapeTexture `json:"CAPE,omitempty"`
}

func NewProfile(id uuid.UUID, name string, modelType ModelType, serializedTextures string) (this Profile) {
	this.Id = id
	this.Name = name
	this.ModelType = modelType
	if len(serializedTextures) < 2 {
		serializedTextures = "{}"
	}
	err := json.Unmarshal([]byte(serializedTextures), &this.Textures)
	if err != nil {
		panic(err)
	}
	return this
}

type ProfileResponse struct {
	Name string `json:"name" binding:"required"`
	Id   string `json:"id" binding:"required"`
}

func (p *Profile) ToSimpleResponse() ProfileResponse {
	return ProfileResponse{
		Id:   util.UnsignedString(p.Id),
		Name: p.Name,
	}
}

func (p *Profile) ToCompleteResponse(signed bool, textureBaseUrl string) (map[string]interface{}, error) {
	textures := TexturesType{}
	if hash, ok := p.Textures["SKIN"]; ok {
		skin := SkinTexture{
			Url: textureBaseUrl + "/" + hash,
		}
		if p.ModelType == ALEX {
			m := MetadataType{
				"model": ALEX,
			}
			skin.Metadata = &m
		}
		textures.SKIN = &skin
	}
	if hash, ok := p.Textures["CAPE"]; ok {
		cape := CapeTexture{
			Url: textureBaseUrl + "/" + hash,
		}
		textures.CAPE = &cape
	}
	texturesStr, err := util.EncodeBase64(util.Property{
		Name: "timestamp", Value: time.Now().UnixMilli(),
	}, util.Property{
		Name: "profileId", Value: util.UnsignedString(p.Id),
	}, util.Property{
		Name: "profileName", Value: p.Name,
	}, util.Property{
		Name: "textures", Value: textures,
	}, util.Property{
		Name: "signatureRequired", Value: signed,
	})
	if err != nil {
		return nil, err
	}
	properties := util.Properties(signed,
		util.StringProperty{Name: "textures", Value: texturesStr},
		util.StringProperty{Name: "uploadableTextures", Value: "skin,cape"},
	)
	return map[string]interface{}{
		"id":         util.UnsignedString(p.Id),
		"name":       p.Name,
		"properties": properties,
	}, nil
}

func (p *Profile) Equals(another *Profile) bool {
	return p == another || p.Id == another.Id
}
