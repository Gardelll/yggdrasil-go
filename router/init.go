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
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
	"yggdrasil-go/service"
)

func InitRouters(router *gin.Engine, db *gorm.DB, meta *ServerMeta, skinRootUrl string) {
	err := router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		panic(err)
	}
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "User-Agent"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	tokenService := service.NewTokenService()
	userService := service.NewUserService(tokenService, db)
	sessionService := service.NewSessionService(tokenService)
	textureService := service.NewTextureService(tokenService, db)
	homeRouter := NewHomeRouter(meta)
	userRouter := NewUserRouter(userService, skinRootUrl)
	sessionRouter := NewSessionRouter(sessionService, skinRootUrl)
	textureRouter := NewTextureRouter(textureService)

	router.GET("/", homeRouter.Home)
	authserver := router.Group("/authserver")
	{
		authserver.POST("/register", userRouter.Register)
		authserver.POST("/authenticate", userRouter.Login)
		authserver.POST("/change", userRouter.ChangeProfile)
		authserver.POST("/refresh", userRouter.Refresh)
		authserver.POST("/validate", userRouter.Validate)
		authserver.POST("/invalidate", userRouter.Invalidate)
		authserver.POST("/signout", userRouter.Signout)
	}
	router.GET("/users/profiles/minecraft/:username", userRouter.UsernameToUUID)
	sessionserver := router.Group("/sessionserver/session/minecraft")
	{
		sessionserver.GET("/profile/:uuid", userRouter.QueryProfile)
		sessionserver.POST("/join", sessionRouter.JoinServer)
		sessionserver.GET("/hasJoined", sessionRouter.HasJoinedServer)
	}
	router.GET("/textures/:hash", textureRouter.GetTexture)
	api := router.Group("/api")
	{
		api.POST("/profiles/minecraft", userRouter.QueryUUIDs)
		api.POST("/user/profile/:uuid/:textureType", textureRouter.SetTexture)
		api.PUT("/user/profile/:uuid/:textureType", textureRouter.UploadTexture)
		api.DELETE("/user/profile/:uuid/:textureType", textureRouter.DeleteTexture)
		api.GET("/users/profiles/minecraft/:username", userRouter.UsernameToUUID)
	}
}
