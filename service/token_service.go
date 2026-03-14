/*
 * Copyright (C) 2022. Gardel <gardel741@outlook.com> and contributors
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
	"time"

	"github.com/google/uuid"
	"yggdrasil-go/cache"
	"yggdrasil-go/model"
	"yggdrasil-go/util"
)

const tokenTTL = 30 * 24 * time.Hour // 30 days, matches Token.GetAvailableLevel() Invalid threshold

type TokenService interface {
	RemoveToken(token *model.Token)
	RemoveAccessToken(accessToken string)
	RemoveAll(profileId uuid.UUID)
	AcquireToken(user *model.User, clientToken *string, profile *model.Profile) *model.Token
	VerifyToken(accessToken string, clientToken *string) model.AvailableLevel
	GetToken(accessToken string) (*model.Token, bool)
	UpdateProfile(profileId uuid.UUID, profile *model.Profile)
}

type tokenStore struct {
	tokenCache cache.IndexedCache[*model.Token]
}

func NewTokenService(tokenCache cache.IndexedCache[*model.Token]) TokenService {
	return &tokenStore{
		tokenCache: tokenCache,
	}
}

func (t *tokenStore) RemoveToken(token *model.Token) {
	t.RemoveAccessToken(token.AccessToken)
}

func (t *tokenStore) RemoveAccessToken(accessToken string) {
	t.tokenCache.Remove(accessToken)
}

func (t *tokenStore) RemoveAll(profileId uuid.UUID) {
	indexKey := util.UnsignedString(profileId)
	t.tokenCache.RemoveByIndex(indexKey)
}

func (t *tokenStore) AcquireToken(user *model.User, clientToken *string, profile *model.Profile) *model.Token {
	if profile == nil {
		var err error
		profile, err = user.Profile()
		if err != nil {
			panic(err)
		}
	}
	token := model.NewToken(util.RandomUUID(), clientToken, profile)
	indexKey := util.UnsignedString(profile.Id)
	_ = t.tokenCache.SetWithIndex(token.AccessToken, &token, tokenTTL, indexKey)
	return &token
}

func (t *tokenStore) VerifyToken(accessToken string, clientToken *string) model.AvailableLevel {
	token, ok := t.tokenCache.Get(accessToken)
	if ok {
		if clientToken != nil && token.ClientToken != *clientToken {
			return model.Invalid
		}
		if token.GetAvailableLevel() == model.Invalid {
			t.RemoveToken(token)
		}
		return token.GetAvailableLevel()
	}
	return model.Invalid
}

func (t *tokenStore) GetToken(accessToken string) (*model.Token, bool) {
	return t.tokenCache.Get(accessToken)
}

func (t *tokenStore) UpdateProfile(profileId uuid.UUID, profile *model.Profile) {
	indexKey := util.UnsignedString(profileId)
	keys := t.tokenCache.GetByIndex(indexKey)
	for _, k := range keys {
		if token, ok := t.tokenCache.Get(k); ok {
			token.SelectedProfile = *profile
			_ = t.tokenCache.SetWithIndex(k, token, tokenTTL, indexKey)
		}
	}
}
