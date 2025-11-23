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

import (
	"time"
)

// Common DTOs shared across multiple modules

// StringProperty represents a name-value-signature property tuple
// Used in profile properties, texture properties, etc.
type StringProperty struct {
	Name      string `json:"name,omitempty"`
	Value     string `json:"value,omitempty"`
	Signature string `json:"signature,omitempty"` // Optional signature field
}

// KeyPair represents a public/private key pair
// Used in profile keys, server keys, etc.
type KeyPair struct {
	PrivateKey string `json:"privateKey,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
}

// Authentication and user management DTOs
// These structures are used for HTTP request/response in authentication endpoints

// RegRequest represents a user registration request
type RegRequest struct {
	Username    string `json:"username" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	ProfileName string `json:"profileName" binding:"required"`
}

// MinecraftAgent represents the Minecraft client agent information
type MinecraftAgent struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

// ClientTokenBase contains the optional client token field
type ClientTokenBase struct {
	ClientToken *string `json:"clientToken,omitempty"`
}

// AccessTokenBase contains the required access token field
type AccessTokenBase struct {
	AccessToken string `json:"accessToken" binding:"required"`
}

// DualTokenBase combines both access token and client token
type DualTokenBase struct {
	AccessTokenBase
	ClientTokenBase
}

// LoginRequest represents a user login request
type LoginRequest struct {
	ClientTokenBase
	Username    string          `json:"username" binding:"required,email"`
	Password    string          `json:"password" binding:"required"`
	RequestUser bool            `json:"requestUser"`
	Agent       *MinecraftAgent `json:"agent,omitempty"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	DualTokenBase
	RequestUser     bool             `json:"requestUser"`
	SelectedProfile *ProfileResponse `json:"selectedProfile,omitempty"`
}

// ValidateRequest represents a token validation request
type ValidateRequest struct {
	DualTokenBase
}

// ChangeProfileRequest represents a profile change request
type ChangeProfileRequest struct {
	DualTokenBase
	ChangeTo string `json:"changeTo" binding:"required"`
}

// InvalidateRequest represents a token invalidation request
type InvalidateRequest struct {
	AccessTokenBase
}

// SignoutRequest represents a user signout request
type SignoutRequest struct {
	Username string `json:"username" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SendEmailRequest represents an email sending request
type SendEmailRequest struct {
	Email     string `json:"email" binding:"required,email"`
	EmailType string `json:"emailType" binding:"required"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	AccessToken string `json:"accessToken" binding:"required"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	User              *UserResponse     `json:"user"`
	ClientToken       string            `json:"clientToken"`
	AccessToken       string            `json:"accessToken"`
	AvailableProfiles []ProfileResponse `json:"availableProfiles,omitempty"`
	SelectedProfile   *ProfileResponse  `json:"selectedProfile"`
}

// UserResponse represents a user information response
type UserResponse struct {
	Username   string           `json:"username,omitempty"`
	Properties []StringProperty `json:"properties"`
	Id         string           `json:"id,omitempty"`
}

// ProfileResponse represents a player profile response
type ProfileResponse struct {
	Name string `json:"name" binding:"required"`
	Id   string `json:"id" binding:"required"`
}

// CompleteProfileResponse represents a complete profile with properties (textures, etc.)
// Used in hasJoined, profile lookup endpoints, and upstream responses
// This is the unified structure for all profile responses with properties
type CompleteProfileResponse struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Properties []StringProperty `json:"properties,omitempty"`
}

// ProfileKeyResponse represents a profile key response
type ProfileKeyResponse struct {
	ExpiresAt            time.Time `json:"expiresAt,omitempty"`
	KeyPair              *KeyPair  `json:"keyPair,omitempty"`
	PublicKeySignature   string    `json:"publicKeySignature,omitempty"`
	PublicKeySignatureV2 string    `json:"publicKeySignatureV2,omitempty"`
	RefreshedAfter       time.Time `json:"refreshedAfter,omitempty"`
}

// ProfileKeyPair is an alias for KeyPair (deprecated, use KeyPair instead)
type ProfileKeyPair = KeyPair

// TexturesType represents the textures field in profile properties
type TexturesType struct {
	Timestamp   int64         `json:"timestamp,omitempty"`
	ProfileID   string        `json:"profileId,omitempty"`
	ProfileName string        `json:"profileName,omitempty"`
	Textures    TexturesValue `json:"textures,omitempty"`
}

// TexturesValue contains skin and cape texture URLs
type TexturesValue struct {
	SKIN *SkinTexture `json:"SKIN,omitempty"`
	CAPE *CapeTexture `json:"CAPE,omitempty"`
}

// MetadataType represents texture metadata
type MetadataType map[string]interface{}

// SkinTexture represents a skin texture URL
type SkinTexture struct {
	Url      string        `json:"url"`
	Metadata *MetadataType `json:"metadata,omitempty"`
}

// CapeTexture represents a cape texture URL
type CapeTexture struct {
	Url string `json:"url"`
}
