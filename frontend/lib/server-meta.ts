/*
 * Copyright (C) 2023-2025. Gardel <gardel741@outlook.com> and contributors
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

import React from 'react'
import axios from 'axios'

export type ServerMeta = {
  meta: {
    serverName?: string
    implementationName?: string
    implementationVersion?: string
  }
}

export function useServerMeta() {
  const [meta, setMeta] = React.useState<ServerMeta | null>(null)
  React.useEffect(() => {
    axios.get<ServerMeta>('/yggdrasil')
      .then(res => {
        setMeta(res.data)
        const name = res.data?.meta?.serverName
        if (name) document.title = name
      })
      .catch(() => {})
  }, [])
  return meta
}
