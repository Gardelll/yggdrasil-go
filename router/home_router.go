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

package router

import (
	"encoding/base64"
	"encoding/pem"
	"github.com/gin-gonic/gin"
	"net/http"
	"yggdrasil-go/dto"
	"yggdrasil-go/service"
)

type HomeRouter interface {
	Home(c *gin.Context)
	PublicKeys(c *gin.Context)
}

type homeRouterImpl struct {
	serverMeta      dto.ServerMeta
	myPubKey        dto.KeyPair
	cachedPubKey    *dto.PublicKeys
	upstreamService service.IUpstreamService // Upstream authentication service (optional)
}

func NewHomeRouter(meta *dto.ServerMeta, upstreamService service.IUpstreamService) HomeRouter {
	signaturePubKey, _ := pem.Decode([]byte(meta.SignaturePublickey))
	homeRouter := homeRouterImpl{
		serverMeta:      *meta,
		myPubKey:        dto.KeyPair{PublicKey: base64.StdEncoding.EncodeToString(signaturePubKey.Bytes)},
		upstreamService: upstreamService,
	}
	return &homeRouter
}

// Home 首页路由
func (h *homeRouterImpl) Home(c *gin.Context) {
	c.JSON(http.StatusOK, h.serverMeta)
}

func (h *homeRouterImpl) PublicKeys(c *gin.Context) {
	if h.cachedPubKey != nil {
		c.JSON(http.StatusOK, h.cachedPubKey)
		return
	}
	publicKeys := dto.PublicKeys{}

	// Use upstream service if configured (tasks.md T018)
	if h.upstreamService != nil {
		upstreamKeys, err := h.upstreamService.GetPublicKeys(c.Request.Context())
		if err == nil && upstreamKeys != nil {
			// Convert upstream public keys to our format (structure now matches directly)
			for _, key := range upstreamKeys.ProfilePropertyKeys {
				publicKeys.ProfilePropertyKeys = append(publicKeys.ProfilePropertyKeys, dto.KeyPair{
					PublicKey: key.PublicKey,
				})
			}
			for _, key := range upstreamKeys.PlayerCertificateKeys {
				publicKeys.PlayerCertificateKeys = append(publicKeys.PlayerCertificateKeys, dto.KeyPair{
					PublicKey: key.PublicKey,
				})
			}
		}
	}

	// Always include our own public key
	publicKeys.ProfilePropertyKeys = append(publicKeys.ProfilePropertyKeys, h.myPubKey)
	publicKeys.PlayerCertificateKeys = append(publicKeys.PlayerCertificateKeys, h.myPubKey)
	c.JSON(http.StatusOK, publicKeys)
	h.cachedPubKey = &publicKeys
}
