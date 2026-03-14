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

import { SkinData } from '@/lib/types'
import { DEFAULT_SKINS, getUUIDHashCode } from '@/lib/skins'
import axios from 'axios'

type ProfileProperty = {
  name: string
  value: string
}

type ProfileTextures = {
  textures?: {
    SKIN?: { url: string; metadata?: { model: string } }
    CAPE?: { url: string }
  }
}

export async function fetchSkinData(uuid: string): Promise<SkinData> {
  const response = await axios.get<{ properties: ProfileProperty[] }>(
    '/sessionserver/session/minecraft/profile/' + uuid
  )
  const texturesProp = response.data.properties.find(v => v.name === 'textures')
  if (!texturesProp) {
    return getDefaultSkin(uuid)
  }

  const profile: ProfileTextures = JSON.parse(window.atob(texturesProp.value))
  const data: SkinData = { skinUrl: '' }

  if (profile.textures?.SKIN) {
    data.skinUrl = profile.textures.SKIN.url
    data.slim = profile.textures.SKIN.metadata?.model === 'slim'
  } else {
    Object.assign(data, getDefaultSkin(uuid))
  }

  if (profile.textures?.CAPE) {
    data.capeUrl = profile.textures.CAPE.url
  }

  return data
}

export function getDefaultSkin(uuid: string): SkinData {
  const index = getUUIDHashCode(uuid) % DEFAULT_SKINS.length
  return { ...DEFAULT_SKINS[index] }
}
