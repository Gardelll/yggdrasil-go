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

package model

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"time"
)

type Texture struct {
	Hash      string `gorm:"size:64;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Data      []byte `gorm:"not null"`
	Used      uint   `gorm:"not null"`
}

func ComputeTextureId(img image.Image) string {
	digest := sha256.New()
	bound := img.Bounds()
	width := bound.Dx()
	height := bound.Dy()
	var buf [4096]byte

	putInt(buf[:], int32(width))
	putInt(buf[4:], int32(height))
	var pos = 8
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			rgba := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if rgba.A == 0 {
				copy(buf[pos:], []byte{0, 0, 0, 0})
			} else {
				copy(buf[pos:], []byte{rgba.A, rgba.R, rgba.G, rgba.B})
			}
			pos += 4
			if pos == len(buf) {
				pos = 0
				digest.Write(buf[:])
			}
		}
	}
	if pos > 0 {
		digest.Write(buf[:pos])
	}
	return hex.EncodeToString(digest.Sum(nil))
}

func putInt(buf []byte, n int32) {
	buf[0] = byte(n >> 24 & 0xff)
	buf[1] = byte(n >> 16 & 0xff)
	buf[2] = byte(n >> 8 & 0xff)
	buf[3] = byte(n & 0xff)
}
