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

package dto

// Server metadata and session DTOs

// MetaInfo represents the server metadata information
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

// ServerMeta represents the complete server metadata response
type ServerMeta struct {
	Meta               MetaInfo `json:"meta"`
	SkinDomains        []string `json:"skinDomains"`
	SignaturePublickey string   `json:"signaturePublickey"`
}

// PublicKeys represents the server public keys response
type PublicKeys struct {
	ProfilePropertyKeys   []KeyPair `json:"profilePropertyKeys,omitempty"`
	PlayerCertificateKeys []KeyPair `json:"playerCertificateKeys,omitempty"`
}

// JoinServerRequest represents a request to join a Minecraft server
type JoinServerRequest struct {
	AccessToken     string `json:"accessToken" binding:"required"`
	SelectedProfile string `json:"selectedProfile" binding:"required"`
	ServerId        string `json:"serverId" binding:"required"`
}

// SetTextureRequest represents a request to set player texture (skin/cape)
type SetTextureRequest struct {
	Url   string `json:"url" binding:"required,url"`
	Model string `json:"model" binding:"oneof=slim default"`
}
