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

package service

import (
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	"yggdrasil-go/model"
	"yggdrasil-go/util"
)

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
	tokenCache *lru.Cache
}

func NewTokenService() TokenService {
	cache, _ := lru.New(10000000)
	store := tokenStore{
		tokenCache: cache,
	}
	return &store
}

func (t *tokenStore) RemoveToken(token *model.Token) {
	t.RemoveAccessToken(token.AccessToken)
}

func (t *tokenStore) RemoveAccessToken(accessToken string) {
	t.tokenCache.Remove(accessToken)
}

func (t *tokenStore) RemoveAll(profileId uuid.UUID) {
	keys := t.tokenCache.Keys()
	for _, k := range keys {
		if v, ok := t.tokenCache.Get(k); ok {
			if v.(*model.Token).SelectedProfile.Id == profileId {
				t.tokenCache.Remove(k)
			}
		}
	}
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
	t.tokenCache.Add(token.AccessToken, &token)
	return &token
}

func (t *tokenStore) VerifyToken(accessToken string, clientToken *string) model.AvailableLevel {
	if value, ok := t.tokenCache.Get(accessToken); ok {
		if token, ok := value.(*model.Token); ok {
			if clientToken != nil && token.ClientToken != *clientToken {
				return model.Invalid
			}
			if token.GetAvailableLevel() == model.Invalid {
				t.RemoveToken(token)
			}
			return token.GetAvailableLevel()
		}
	}
	return model.Invalid
}

func (t *tokenStore) GetToken(accessToken string) (*model.Token, bool) {
	if value, ok := t.tokenCache.Get(accessToken); ok {
		if token, ok := value.(*model.Token); ok {
			return token, true
		}
	}
	return nil, false
}

func (t *tokenStore) UpdateProfile(profileId uuid.UUID, profile *model.Profile) {
	keys := t.tokenCache.Keys()
	for _, k := range keys {
		if v, ok := t.tokenCache.Get(k); ok {
			if token := v.(*model.Token); token.SelectedProfile.Id == profileId {
				token.SelectedProfile = *profile
			}
		}
	}
}
