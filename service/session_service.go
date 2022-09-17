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
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"net/http"
	"net/url"
	"yggdrasil-go/model"
	"yggdrasil-go/util"
)

type SessionService interface {
	JoinServer(accessToken string, serverId string, selectedProfile string, ip string) error
	HasJoinedServer(serverId string, username string, ip string, textureBaseUrl string) (map[string]interface{}, error)
}

type sessionStore struct {
	sessionCache *lru.Cache
	tokenService TokenService
}

func NewSessionService(service TokenService) SessionService {
	cache, _ := lru.New(100000)
	store := sessionStore{
		sessionCache: cache,
		tokenService: service,
	}
	return &store
}

func (s *sessionStore) JoinServer(accessToken string, serverId string, selectedProfile string, ip string) error {
	token, ok := s.tokenService.GetToken(accessToken)
	if ok {
		if token.GetAvailableLevel() != model.Valid ||
			util.UnsignedString(token.SelectedProfile.Id) != selectedProfile {
			return util.NewForbiddenOperationError(util.MessageInvalidToken)
		}
		session := model.NewAuthenticationSession(serverId, token, ip)
		s.sessionCache.Add(serverId, &session)
	} else {
		data := map[string]string{
			"accessToken":     accessToken,
			"selectedProfile": selectedProfile,
			"serverId":        serverId,
		}
		err := util.PostObjectForError("https://sessionserver.mojang.com/session/minecraft/join", data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *sessionStore) HasJoinedServer(serverId string, username string, ip string, textureBaseUrl string) (map[string]interface{}, error) {
	if value, ok := s.sessionCache.Get(serverId); ok {
		if session, ok := value.(*model.AuthenticationSession); ok {
			if !(session.HasExpired() && s.sessionCache.Remove(serverId)) &&
				(ip == "" || ip == session.Ip) && (session.Token.SelectedProfile.Name == username) {
				return session.Token.SelectedProfile.ToCompleteResponse(true, textureBaseUrl)
			}
		}
	} else {
		m := make(map[string]interface{})
		includeIp := ""
		if ip != "" {
			includeIp = "&ip=" + url.QueryEscape(ip)
		}
		err := util.GetObject(fmt.Sprintf("https://sessionserver.mojang.com/session/minecraft/hasJoined?username=%s&serverId=%s%s", url.QueryEscape(username), url.QueryEscape(serverId), includeIp), &m)
		if err != nil {
			return nil, err
		} else {
			return m, nil
		}
	}
	return nil, util.YggdrasilError{Status: http.StatusNoContent}
}
