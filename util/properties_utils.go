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

package util

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"log"
)

type Property struct {
	Name  string
	Value interface{}
}

type StringProperty struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// PrivateKey RSA PKCS8 Private Key
var PrivateKey *rsa.PrivateKey

func EncodeBase64(properties ...Property) (string, error) {
	obj := make(map[string]interface{})
	for _, property := range properties {
		obj[property.Name] = property.Value
	}
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

func Properties(sign bool, properties ...StringProperty) []map[string]string {
	list := make([]map[string]string, 0, len(properties))
	for _, property := range properties {
		obj := map[string]string{
			"name":  property.Name,
			"value": property.Value,
		}
		if sign {
			err := error(nil)
			obj["signature"], err = Sign(property.Value)
			if err != nil {
				log.Printf("无法签名字符串 '%s', 原因: %s", property.Value, err.Error())
			}
		}
		list = append(list, obj)
	}
	return list
}

func Sign(value string) (string, error) {
	if PrivateKey == nil {
		panic("未初始化私钥")
	}
	sum := sha1.Sum([]byte(value))
	sig, err := rsa.SignPKCS1v15(rand.Reader, PrivateKey, crypto.SHA1, sum[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}
