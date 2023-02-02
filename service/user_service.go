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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"yggdrasil-go/model"
	"yggdrasil-go/util"
)

type UserService interface {
	Register(username string, password string, profileName string) (*model.UserResponse, error)
	Login(username string, password string, clientToken *string, requestUser bool) (*LoginResponse, error)
	ChangeProfile(accessToken string, clientToken *string, changeTo string) error
	Refresh(accessToken string, clientToken *string, requestUser bool, selectedProfile *model.ProfileResponse) (*LoginResponse, error)
	Validate(accessToken string, clientToken *string) error
	Invalidate(accessToken string) error
	Signout(username string, password string) error
	UsernameToUUID(username string) (*model.ProfileResponse, error)
	QueryUUIDs(usernames []string) ([]model.ProfileResponse, error)
	QueryProfile(profileId uuid.UUID, unsigned bool, textureBaseUrl string) (map[string]interface{}, error)
	ProfileKey(accessToken string) (*ProfileKeyResponse, error)
}

type LoginResponse struct {
	User              *model.UserResponse     `json:"user"`
	ClientToken       string                  `json:"clientToken"`
	AccessToken       string                  `json:"accessToken"`
	AvailableProfiles []model.ProfileResponse `json:"availableProfiles,omitempty"`
	SelectedProfile   *model.ProfileResponse  `json:"selectedProfile"`
}

type ProfileKeyResponse struct {
	ExpiresAt            time.Time       `json:"expiresAt,omitempty"`
	KeyPair              *ProfileKeyPair `json:"keyPair,omitempty"`
	PublicKeySignature   string          `json:"publicKeySignature,omitempty"`
	PublicKeySignatureV2 string          `json:"publicKeySignatureV2,omitempty"`
	RefreshedAfter       time.Time       `json:"refreshedAfter,omitempty"`
}

type ProfileKeyPair struct {
	PrivateKey string `json:"privateKey,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
}

type userServiceImpl struct {
	tokenService    TokenService
	db              *gorm.DB
	limitLruCache   *lru.Cache
	profileKeyCache *lru.Cache
	keyPairCh       chan ProfileKeyPair
}

func NewUserService(tokenService TokenService, db *gorm.DB) UserService {
	cache0, _ := lru.New(10000)
	cache1, _ := lru.New(10000)
	ch := make(chan ProfileKeyPair, 100)
	userService := userServiceImpl{
		tokenService:    tokenService,
		db:              db,
		limitLruCache:   cache0,
		profileKeyCache: cache1,
		keyPairCh:       ch,
	}
	go userService.genKeyPair()
	return &userService
}

func (u *userServiceImpl) Register(username string, password string, profileName string) (*model.UserResponse, error) {
	var count int64
	if err := u.db.Table("users").Where("email = ?", username).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, util.NewForbiddenOperationError("email exist")
	}
	if err := u.db.Table("users").Where("profile_name = ?", profileName).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, util.NewForbiddenOperationError("profileName exist")
	} else if _, err := mojangUsernameToUUID(profileName); err == nil {
		return nil, util.NewForbiddenOperationError("profileName duplicate")
	}
	matched, err := regexp.MatchString("^(\\w){3,}(\\.\\w+)*@(\\w){2,}((\\.\\w+)+)$", username)
	if err != nil {
		return nil, err
	}
	if !matched || len(password) < 6 || isInvalidProfileName(profileName) {
		return nil, util.NewIllegalArgumentError("bad format(valid email, password longer than 5, profileName longer than 1)")
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := model.User{
		ID:       uuid.New(),
		Email:    username,
		Password: string(hashedPass),
	}
	profile := model.NewProfile(user.ID, profileName, model.STEVE, "")
	user.SetProfile(&profile)

	if err := u.db.Create(&user).Error; err != nil {
		return nil, err
	}
	response := user.ToResponse()
	return &response, nil
}

func isInvalidProfileName(name string) bool {
	// To support Unicode (like Chinese) profile name, abandoned treatment.
	return name == "" || strings.ContainsRune(name, ' ') || len(name) <= 1
	//return name == "" || !name.matches("^[0-1a-zA-Z_]{2,16}$");
}

func (u *userServiceImpl) Login(username string, password string, clientToken *string, requestUser bool) (*LoginResponse, error) {
	if !u.allowUser(username) {
		return nil, util.YggdrasilError{
			Status:       http.StatusTooManyRequests,
			ErrorCode:    "ForbiddenOperationException",
			ErrorMessage: "Forbidden",
		}
	}
	user := model.User{}
	if err := u.db.Where("email = ?", username).First(&user).Error; err == nil {
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil {
			var useClientToken string
			if clientToken == nil || *clientToken == "" {
				useClientToken = util.RandomUUID()
			} else {
				useClientToken = *clientToken
			}
			token := u.tokenService.AcquireToken(&user, &useClientToken, nil)
			profile, err := user.Profile()
			if err != nil {
				panic(err)
			}
			simpleResponse := profile.ToSimpleResponse()
			var response = LoginResponse{
				AccessToken:       token.AccessToken,
				ClientToken:       token.ClientToken,
				AvailableProfiles: []model.ProfileResponse{simpleResponse},
				SelectedProfile:   &simpleResponse,
			}
			userResponse := user.ToResponse()
			if requestUser {
				response.User = &userResponse
			}
			return &response, nil
		}
	}

	return nil, util.NewForbiddenOperationError(util.MessageInvalidCredentials)
}

func (u *userServiceImpl) ChangeProfile(accessToken string, clientToken *string, changeTo string) error {
	if u.tokenService.VerifyToken(accessToken, clientToken) != model.Valid {
		return util.NewForbiddenOperationError(util.MessageInvalidToken)
	}
	token, ok := u.tokenService.GetToken(accessToken)
	if !ok {
		return util.NewForbiddenOperationError(util.MessageInvalidToken)
	}
	user := model.User{}
	profile := token.SelectedProfile
	err := u.db.First(&user, profile.Id).Error
	if err != nil {
		return util.NewForbiddenOperationError("User not found")
	}
	var count int64
	if err := u.db.Table("users").Where("profile_name = ?", changeTo).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return util.NewForbiddenOperationError("profileName exist")
	} else if _, err := mojangUsernameToUUID(changeTo); err == nil {
		return util.NewForbiddenOperationError("profileName duplicate")
	}
	if isInvalidProfileName(changeTo) {
		return util.NewForbiddenOperationError("bad format(profileName longer than 1)")
	}

	if err = u.db.Model(&user).Update("profile_name", changeTo).Error; err != nil {
		return err
	}
	profile.Name = changeTo
	u.tokenService.UpdateProfile(user.ID, &profile)
	return nil
}

func (u *userServiceImpl) Refresh(accessToken string, clientToken *string, requestUser bool, selectedProfile *model.ProfileResponse) (*LoginResponse, error) {
	if len(accessToken) <= 36 {
		user := model.User{}
		if selectedProfile != nil {
			// 由于当前实现把用户 UUID 作为角色 UUID，所以不支持角色选择，只要选择了就会报错
			return nil, util.NewForbiddenOperationError(util.MessageTokenAlreadyAssigned)
		}
		if u.tokenService.VerifyToken(accessToken, clientToken) == model.Invalid {
			return nil, util.NewForbiddenOperationError(util.MessageInvalidToken)
		}
		token, ok := u.tokenService.GetToken(accessToken)
		if !ok {
			return nil, util.NewForbiddenOperationError(util.MessageInvalidToken)
		}

		if err := u.db.First(&user, token.SelectedProfile.Id).Error; err != nil {
			return nil, util.NewIllegalArgumentError(util.MessageProfileNotFound)
		}
		newToken := u.tokenService.AcquireToken(&user, clientToken, nil)
		u.tokenService.RemoveAccessToken(accessToken)
		simpleResponse := newToken.SelectedProfile.ToSimpleResponse()
		var response = LoginResponse{
			AccessToken:       newToken.AccessToken,
			ClientToken:       newToken.ClientToken,
			AvailableProfiles: []model.ProfileResponse{},
			SelectedProfile:   &simpleResponse,
		}
		userResponse := user.ToResponse()
		if requestUser {
			response.User = &userResponse
		}
		return &response, nil
	} else {
		data := map[string]interface{}{
			"accessToken":     accessToken,
			"clientToken":     clientToken,
			"requestUser":     requestUser,
			"selectedProfile": selectedProfile,
		}
		loginResponse := LoginResponse{}
		err := util.PostObject("https://authserver.mojang.com/refresh", data, &loginResponse)
		if err != nil {
			return nil, err
		} else {
			return &loginResponse, nil
		}
	}
}

func (u *userServiceImpl) Validate(accessToken string, clientToken *string) error {
	if len(accessToken) <= 36 {
		if u.tokenService.VerifyToken(accessToken, clientToken) != model.Valid {
			return util.NewForbiddenOperationError(util.MessageInvalidToken)
		} else {
			return nil
		}
	} else {
		data := map[string]interface{}{
			"accessToken": accessToken,
			"clientToken": clientToken,
		}
		err := util.PostObjectForError("https://authserver.mojang.com/validate", data)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (u *userServiceImpl) Invalidate(accessToken string) error {
	if len(accessToken) <= 36 {
		u.tokenService.RemoveAccessToken(accessToken)
	} else {
		data := map[string]interface{}{
			"accessToken": accessToken,
		}
		err := util.PostObjectForError("https://authserver.mojang.com/invalidate", data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *userServiceImpl) Signout(username string, password string) error {
	if !u.allowUser(username) {
		return util.YggdrasilError{
			Status:       http.StatusTooManyRequests,
			ErrorCode:    "ForbiddenOperationException",
			ErrorMessage: "Forbidden",
		}
	}
	user := model.User{}
	if err := u.db.Where("email = ?", username).First(&user).Error; err == nil {
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil {
			u.tokenService.RemoveAll(user.ID)
			return nil
		} else {
			return util.NewForbiddenOperationError(util.MessageInvalidCredentials)
		}
	} else {
		data := map[string]interface{}{
			"username": username,
			"password": password,
		}
		err := util.PostObjectForError("https://authserver.mojang.com/signout", data)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (u *userServiceImpl) UsernameToUUID(username string) (*model.ProfileResponse, error) {
	user := model.User{}
	if result := u.db.Where("profile_name = ?", username).First(&user); result.Error == nil {
		return &model.ProfileResponse{
			Name: user.ProfileName,
			Id:   util.UnsignedString(user.ID),
		}, nil
	} else {
		response, err := mojangUsernameToUUID(username)
		if err != nil {
			return nil, nil
		} else {
			return &response, nil
		}
	}
}

func (u *userServiceImpl) QueryUUIDs(usernames []string) ([]model.ProfileResponse, error) {
	var users []model.User
	var names []string
	if len(usernames) > 10 {
		names = usernames[:10]
	} else {
		names = usernames
	}
	var responses = make([]model.ProfileResponse, 0)
	if err := u.db.Table("users").Where("profile_name in ?", names).Find(&users).Error; err == nil {
		for _, user := range users {
			responses = append(responses, model.ProfileResponse{
				Name: user.ProfileName,
				Id:   util.UnsignedString(user.ID),
			})
		}
	}
	return responses, nil
}

func (u *userServiceImpl) QueryProfile(profileId uuid.UUID, unsigned bool, textureBaseUrl string) (map[string]interface{}, error) {
	user := model.User{}
	if err := u.db.First(&user, profileId).Error; err == nil {
		profile, err := user.Profile()
		if err != nil {
			return nil, err
		}
		response, err := profile.ToCompleteResponse(!unsigned, textureBaseUrl)
		if err != nil {
			return nil, err
		} else {
			return response, err
		}
	} else {
		result := map[string]interface{}{}
		err := util.GetObject(fmt.Sprintf("https://sessionserver.mojang.com/session/minecraft/profile/%s?unsigned=%t", util.UnsignedString(profileId), unsigned), &result)
		if err != nil {
			return nil, err
		} else {
			return result, nil
		}
	}
}

func (u *userServiceImpl) ProfileKey(accessToken string) (resp *ProfileKeyResponse, err error) {
	token, ok := u.tokenService.GetToken(accessToken)
	if ok && token.GetAvailableLevel() == model.Valid {
		resp = new(ProfileKeyResponse)
		now := time.Now().UTC()
		resp.RefreshedAfter = now
		resp.ExpiresAt = now.Add(time.Hour * 24 * 90)
		keyPair, err := u.getProfileKey(token.SelectedProfile.Id)
		if err != nil {
			return nil, err
		}
		resp.KeyPair = keyPair
		signStr := fmt.Sprintf("%d%s", resp.ExpiresAt.UnixMilli(), keyPair.PublicKey)
		sign, err := util.Sign(signStr)
		if err != nil {
			return nil, err
		}
		resp.PublicKeySignature = sign
		resp.PublicKeySignatureV2 = sign
	} else {
		err = util.PostForString("https://api.minecraftservices.com/player/certificates", accessToken, []byte(""), resp)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func (u *userServiceImpl) allowUser(username string) bool {
	if value, ok := u.limitLruCache.Get(username); ok {
		if limiter, ok := value.(*rate.Limiter); ok {
			return limiter.Allow()
		} else {
			u.limitLruCache.Remove(username)
		}
	} else {
		limiter := rate.NewLimiter(0.2, 3)
		u.limitLruCache.Add(username, limiter)
	}
	return true
}

func (u *userServiceImpl) getProfileKey(profileId uuid.UUID) (*ProfileKeyPair, error) {
	if value, ok := u.profileKeyCache.Get(profileId); ok {
		if keyPair, ok := value.(*ProfileKeyPair); ok {
			return keyPair, nil
		}
	}
	if keyPair, ok := <-u.keyPairCh; ok {
		u.profileKeyCache.Add(profileId, &keyPair)
		return &keyPair, nil
	} else {
		return nil, errors.New("unable to generate rsa key pair")
	}
}

func (u *userServiceImpl) genKeyPair() {
	for {
		keyPair := ProfileKeyPair{}
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			close(u.keyPairCh)
			panic(err)
		}
		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			close(u.keyPairCh)
			panic(err)
		}
		keyPair.PrivateKey = string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		}))
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			close(u.keyPairCh)
			panic(err)
		}
		keyPair.PublicKey = string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyBytes,
		}))
		u.keyPairCh <- keyPair
	}
}

func mojangUsernameToUUID(username string) (model.ProfileResponse, error) {
	response := model.ProfileResponse{}
	reqUrl := fmt.Sprintf("https://api.mojang.com/users/profiles/minecraft/%s", url.PathEscape(username))
	err := util.GetObject(reqUrl, &response)
	if err != nil {
		return response, err
	} else {
		return response, nil
	}
}
