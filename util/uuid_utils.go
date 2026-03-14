/*
 * Copyright (C) 2022. Gardel <gardel741@outlook.com> and contributors
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
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
)

func UnsignedString(u uuid.UUID) string {
	var buf [32]byte
	hex.Encode(buf[:], u[:])
	return string(buf[:])
}

func ToUUID(str string) (uuid.UUID, error) {
	return uuid.Parse(str)
}

func RandomUUID() string {
	return UnsignedString(uuid.New())
}

// OfflineUUIDFromName computes a deterministic UUID from a player name,
// compatible with Java's UUID.nameUUIDFromBytes(("OfflinePlayer:" + name).getBytes(UTF_8)).
// This is NOT a standard RFC 4122 namespace-based UUID v3; it hashes the raw
// "OfflinePlayer:<name>" bytes with MD5, then sets version=3 and IETF variant bits.
func OfflineUUIDFromName(name string) uuid.UUID {
	hash := md5.Sum([]byte("OfflinePlayer:" + name))
	hash[6] = (hash[6] & 0x0f) | 0x30 // version 3
	hash[8] = (hash[8] & 0x3f) | 0x80 // IETF variant
	return uuid.UUID(hash)
}
