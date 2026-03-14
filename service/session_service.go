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

package service

import (
	"context"
	"net/http"
	"time"
	"yggdrasil-go/cache"
	"yggdrasil-go/dto"
	"yggdrasil-go/model"
	"yggdrasil-go/util"
)

const sessionTTL = 30 * time.Second // matches AuthenticationSession.HasExpired()

type SessionService interface {
	JoinServer(ctx context.Context, accessToken string, serverId string, selectedProfile string, ip string) error
	HasJoinedServer(ctx context.Context, serverId string, username string, ip string, textureBaseUrl string) (*dto.CompleteProfileResponse, error)
}

type sessionStore struct {
	sessionCache    cache.Cache[*model.AuthenticationSession]
	tokenService    TokenService
	upstreamService IUpstreamService // Upstream authentication service (optional)
}

func NewSessionService(sessionCache cache.Cache[*model.AuthenticationSession], service TokenService, upstreamService IUpstreamService) SessionService {
	return &sessionStore{
		sessionCache:    sessionCache,
		tokenService:    service,
		upstreamService: upstreamService,
	}
}

func (s *sessionStore) JoinServer(ctx context.Context, accessToken string, serverId string, selectedProfile string, ip string) error {
	token, ok := s.tokenService.GetToken(accessToken)
	if ok {
		if token.GetAvailableLevel() != model.Valid ||
			util.UnsignedString(token.SelectedProfile.Id) != selectedProfile {
			return util.NewForbiddenOperationError(util.MessageInvalidToken)
		}
		session := model.NewAuthenticationSession(serverId, token, ip)
		_ = s.sessionCache.Set(serverId, &session, sessionTTL)
	} else {
		// Use upstream service if configured (tasks.md T018)
		// Forward the join request to upstream authentication service
		// Pass all parameters from the original API request
		if s.upstreamService != nil {
			err := s.upstreamService.VerifySession(ctx, accessToken, selectedProfile, serverId)
			if err != nil {
				return err
			}
		}
		// Note: In degraded mode (no upstream), session verification is skipped
	}
	return nil
}

func (s *sessionStore) HasJoinedServer(ctx context.Context, serverId string, username string, ip string, textureBaseUrl string) (*dto.CompleteProfileResponse, error) {
	if session, ok := s.sessionCache.Get(serverId); ok {
		if !(session.HasExpired() && s.sessionCache.Remove(serverId)) &&
			(ip == "" || ip == session.Ip) && (session.Token.SelectedProfile.Name == username) {
			return session.Token.SelectedProfile.ToCompleteResponse(true, textureBaseUrl)
		}
	} else {
		// Use upstream service if configured (tasks.md T018)
		if s.upstreamService != nil {
			var ipPtr *string
			if ip != "" {
				ipPtr = &ip
			}
			joinedResp, err := s.upstreamService.HasJoined(ctx, username, serverId, ipPtr)
			if err != nil {
				return nil, err
			}
			if joinedResp != nil {
				// JoinedResponse is now an alias of CompleteProfileResponse
				return joinedResp, nil
			}
		}
		// Degraded mode: no upstream configured
		return nil, util.YggdrasilError{Status: http.StatusNoContent}
	}
	return nil, util.YggdrasilError{Status: http.StatusNoContent}
}
