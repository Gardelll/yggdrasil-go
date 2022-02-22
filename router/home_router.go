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
)

type MetaInfo struct {
	ImplementationName    string `json:"implementationName,omitempty"`
	ImplementationVersion string `json:"implementationVersion,omitempty"`
	ServerName            string `json:"serverName,omitempty"`
	Links                 struct {
		Homepage string `json:"homepage,omitempty"`
		Register string `json:"register,omitempty"`
	} `json:"links"`
	FeatureNonEmailLogin     bool `json:"feature.non_email_login,omitempty"`
	FeatureLegacySkinApi     bool `json:"feature.legacy_skin_api,omitempty"`
	FeatureNoMojangNamespace bool `json:"feature.no_mojang_namespace,omitempty"`
}

type ServerMeta struct {
	Meta               MetaInfo `json:"meta"`
	SkinDomains        []string `json:"skinDomains"`
	SignaturePublickey string   `json:"signaturePublickey"`
}

type HomeRouter interface {
	Home(c *gin.Context)
}

type homeRouterImpl struct {
	serverMeta ServerMeta
}

func NewHomeRouter(meta *ServerMeta) HomeRouter {
	homeRouter := homeRouterImpl{
		serverMeta: *meta,
	}
	return &homeRouter
}

// Home 首页路由
func (h *homeRouterImpl) Home(c *gin.Context) {
	c.JSON(http.StatusOK, h.serverMeta)
}
