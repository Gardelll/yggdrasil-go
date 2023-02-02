/*
 * Copyright (C) 2022-2023. Gardel <sunxinao@hotmail.com> and contributors
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

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"errors"
	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
	"gorm.io/gorm"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yggdrasil-go/model"
	"yggdrasil-go/router"
	"yggdrasil-go/util"
)

//go:embed assets/*
var f embed.FS

type MetaCfg struct {
	ServerName            string   `ini:"server_name"`
	ImplementationName    string   `ini:"implementation_name"`
	ImplementationVersion string   `ini:"implementation_version"`
	SkinDomains           []string `ini:"skin_domains"`
	SkinRootUrl           string   `ini:"skin_root_url"`
}

type ServerCfg struct {
	ServerAddress  string   `ini:"server_address"`
	TrustedProxies []string `ini:"trusted_proxies"`
}

func main() {
	configFilePath := "config.ini"
	cfg, err := ini.LooseLoad(configFilePath)
	if err != nil {
		log.Fatal("无法读取配置文件", err)
	}
	meta := MetaCfg{
		ServerName:            "A Mojang Yggdrasil Server",
		ImplementationName:    "go-yggdrasil-server",
		ImplementationVersion: "v0.1",
		SkinDomains:           []string{".example.com", "localhost"},
		SkinRootUrl:           "http://localhost:8080",
	}
	err = cfg.Section("meta").MapTo(&meta)
	if err != nil {
		log.Fatal("无法读取配置文件", err)
	}
	dbCfg := util.DbCfg{
		DatabaseDriver: "sqlite",
		DatabaseDsn:    "file:sqlite.db?cache=shared",
	}
	err = cfg.Section("database").MapTo(&dbCfg)
	if err != nil {
		log.Fatal("无法读取配置文件", err)
	}
	pathSection := cfg.Section("paths")
	privateKeyPath := pathSection.Key("private_key_file").MustString("private.pem")
	publicKeyPath := pathSection.Key("public_key_file").MustString("public.pem")
	serverCfg := ServerCfg{
		ServerAddress: ":8080",
		TrustedProxies: []string{
			"127.0.0.0/8",
			"10.0.0.0/8",
			"192.168.0.0/16",
			"172.16.0.0/12",
		},
	}
	err = cfg.Section("server").MapTo(&serverCfg)
	if err != nil {
		log.Fatal("无法读取配置文件", err)
	}
	_, err = os.Stat(configFilePath)
	if err != nil && os.IsNotExist(err) {
		log.Println("配置文件不存在，已使用默认配置")
		_ = cfg.Section("meta").ReflectFrom(&meta)
		_ = cfg.Section("database").ReflectFrom(&dbCfg)
		_ = cfg.Section("server").ReflectFrom(&serverCfg)
		err = cfg.SaveToIndent(configFilePath, " ")
		if err != nil {
			log.Println("警告: 无法保存配置文件", err)
		}
	}
	checkRsaKeyFile(privateKeyPath, publicKeyPath)
	publicKeyContent, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal("无法读取公钥内容", err)
	}
	db, err := gorm.Open(util.GetDialector(dbCfg), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatal("无法连接数据库", err)
	}
	err = db.AutoMigrate(&model.User{}, &model.Texture{})
	if err != nil {
		log.Fatal("无法导入数据库", err)
	}
	serverMeta := router.ServerMeta{}
	serverMeta.Meta.ServerName = meta.ServerName
	serverMeta.Meta.ImplementationName = meta.ImplementationName
	serverMeta.Meta.ImplementationVersion = meta.ImplementationVersion
	serverMeta.Meta.FeatureNoMojangNamespace = true
	serverMeta.Meta.FeatureEnableProfileKey = true
	serverMeta.Meta.Links.Homepage = meta.SkinRootUrl + "/profile/"
	serverMeta.Meta.Links.Register = meta.SkinRootUrl + "/profile/"
	serverMeta.SkinDomains = meta.SkinDomains
	serverMeta.SignaturePublickey = string(publicKeyContent)
	r := gin.Default()
	err = r.SetTrustedProxies(serverCfg.TrustedProxies)
	if err != nil {
		log.Fatal(err)
	}
	router.InitRouters(r, db, &serverMeta, meta.SkinRootUrl)
	assetsFs, err := fs.Sub(f, "assets")
	if err != nil {
		log.Fatal(err)
	}
	r.StaticFS("/profile", http.FS(assetsFs))
	srv := &http.Server{
		Addr:    serverCfg.ServerAddress,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()
	log.Printf("已启动, 地址: %s\n", srv.Addr)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("关闭...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("强制关闭:", err)
	}
	log.Println("退出")
}

func checkRsaKeyFile(privateKeyPath string, publicKeyPath string) {
	_, err := os.Stat(privateKeyPath)
	if err != nil && os.IsNotExist(err) {
		privatePem, err := os.OpenFile(privateKeyPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatalln("无法创建私钥文件", err)
		}
		defer privatePem.Close()
		publicPem, err := os.OpenFile(publicKeyPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalln("无法创建公钥文件", err)
		}
		defer publicPem.Close()
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		util.PrivateKey = privateKey
		if err != nil {
			log.Fatalln("无法生成 RSA 密钥", err)
		}
		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			log.Fatalln("无法序列化 RSA 密钥", err)
		}
		err = pem.Encode(privatePem, &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateKeyBytes,
		})
		if err != nil {
			log.Fatalln("无法写入私钥文件", err)
		}
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			log.Fatalln("无法序列化 RSA 公钥", err)
		}
		err = pem.Encode(publicPem, &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		})
		if err != nil {
			log.Fatalln("无法写入公钥文件", err)
		}
	} else if err != nil {
		log.Fatalln("无法打开私钥文件", err)
	} else {
		pemContent, err := os.ReadFile(privateKeyPath)
		if err != nil {
			log.Fatalln("无法打开私钥文件", err)
		}
		pemBlock, _ := pem.Decode(pemContent)
		privateKeyI, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			log.Fatalln("无法解析私钥文件", err)
		}
		util.PrivateKey = privateKeyI.(*rsa.PrivateKey)
	}
}
