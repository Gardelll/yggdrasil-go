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
	"yggdrasil-go/util"
)

type MetaInfo struct {
	ImplementationName    string `json:"implementationName,omitempty"`
	ImplementationVersion string `json:"implementationVersion,omitempty"`
	ServerName            string `json:"serverName,omitempty"`
	Links                 struct {
		Homepage string `json:"homepage,omitempty"`
		Register string `json:"register,omitempty"`
	} `json:"links"`
	FeatureNonEmailLogin            bool `json:"feature.non_email_login,omitempty"`
	FeatureLegacySkinApi            bool `json:"feature.legacy_skin_api,omitempty"`
	FeatureNoMojangNamespace        bool `json:"feature.no_mojang_namespace,omitempty"`
	FeatureEnableProfileKey         bool `json:"feature.enable_profile_key,omitempty"`
	FeatureEnableMojangAntiFeatures bool `json:"feature.enable_mojang_anti_features,omitempty"`
}

type ServerMeta struct {
	Meta               MetaInfo `json:"meta"`
	SkinDomains        []string `json:"skinDomains"`
	SignaturePublickey string   `json:"signaturePublickey"`
}

type KeyPair struct {
	PrivateKey string `json:"privateKey,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
}

type PublicKeys struct {
	ProfilePropertyKeys   []KeyPair `json:"profilePropertyKeys,omitempty"`
	PlayerCertificateKeys []KeyPair `json:"playerCertificateKeys,omitempty"`
}

type HomeRouter interface {
	Home(c *gin.Context)
	PublicKeys(c *gin.Context)
}

type homeRouterImpl struct {
	serverMeta   ServerMeta
	myPubKey     KeyPair
	cachedPubKey *PublicKeys
}

func NewHomeRouter(meta *ServerMeta) HomeRouter {
	signaturePubKey, _ := pem.Decode([]byte(meta.SignaturePublickey))
	homeRouter := homeRouterImpl{
		serverMeta: *meta,
		myPubKey:   KeyPair{PublicKey: base64.StdEncoding.EncodeToString(signaturePubKey.Bytes)},
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
	publicKeys := PublicKeys{}
	err := util.GetObject("https://api.minecraftservices.com/publickeys", &publicKeys)
	if err != nil {
		util.HandleError(c, err)
		return
	}
	publicKeys.ProfilePropertyKeys = append(publicKeys.ProfilePropertyKeys, h.myPubKey)
	publicKeys.PlayerCertificateKeys = append(publicKeys.PlayerCertificateKeys, h.myPubKey)
	c.JSON(http.StatusOK, publicKeys)
	h.cachedPubKey = &publicKeys
}
