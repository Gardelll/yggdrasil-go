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

package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
	"yggdrasil-go/dto"
	"yggdrasil-go/util"
)

type User struct {
	ID                 uuid.UUID `gorm:"column:id;type:string;size:36;primaryKey"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Email              string   `gorm:"size:64;uniqueIndex:email_idx"`
	Password           string   `gorm:"size:255"`
	EmailVerified      bool     `gorm:"default:false"`
	ProfileName        string   `gorm:"size:64;uniqueIndex:profile_name_idx"`
	ProfileModelType   string   `gorm:"size:8;default:STEVE"`
	SerializedTextures string   `gorm:"type:TEXT NULL"`
	profile            *Profile `gorm:"-"`
}

func (u *User) Profile() (*Profile, error) {
	if len(u.ProfileName) == 0 {
		return nil, util.NewIllegalArgumentError("Do not have profile")
	}
	if u.profile != nil {
		return u.profile, nil
	} else {
		var modelType ModelType
		if u.ProfileModelType == "ALEX" {
			modelType = ALEX
		} else {
			modelType = STEVE
		}
		profile := NewProfile(u.ID, u.ProfileName, modelType, u.SerializedTextures)
		return &profile, nil
	}
}

func (u *User) SetProfile(p *Profile) {
	u.profile = p
	u.ProfileName = p.Name
	switch p.ModelType {
	case ALEX:
		u.ProfileModelType = "ALEX"
		break
	case STEVE:
		u.ProfileModelType = "STEVE"
		break
	}
	serialized, err := json.Marshal(p.Textures)
	if err != nil {
		panic("Can not serialize texture")
	}
	u.SerializedTextures = string(serialized)
}

// UserResponse moved to dto package
type UserResponse = dto.UserResponse

func (u *User) ToResponse() dto.UserResponse {
	return UserResponse{
		Id:         util.UnsignedString(u.ID),
		Username:   u.ProfileName,
		Properties: make([]dto.StringProperty, 0),
	}
}
