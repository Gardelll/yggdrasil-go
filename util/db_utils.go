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
	"gorm.io/gorm"
	"log"
	"strings"
	"yggdrasil-go/util/dialector"
)

type DbCfg struct {
	DatabaseDriver string `ini:"database_driver"`
	DatabaseDsn    string `ini:"database_dsn"`
}

func GetDialector(cfg DbCfg) gorm.Dialector {
	if driver, ok := dialector.DbDriverDialectors[cfg.DatabaseDriver]; ok {
		return driver(cfg.DatabaseDsn)
	} else {
		keys := make([]string, len(dialector.DbDriverDialectors))
		p := 0
		for k := range dialector.DbDriverDialectors {
			keys[p] = k
			p++
		}
		supported := strings.Join(keys, ",")
		log.Panicf("Unknown driver: %s\nSupported: %s\n", cfg.DatabaseDriver, supported)
		return nil
	}
}
