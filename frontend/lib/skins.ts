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

export const DEFAULT_SKINS: SkinData[] = [
  { skinUrl: '/profile/player/slim/alex.png', slim: true },
  { skinUrl: '/profile/player/slim/ari.png', slim: true },
  { skinUrl: '/profile/player/slim/efe.png', slim: true },
  { skinUrl: '/profile/player/slim/kai.png', slim: true },
  { skinUrl: '/profile/player/slim/makena.png', slim: true },
  { skinUrl: '/profile/player/slim/noor.png', slim: true },
  { skinUrl: '/profile/player/slim/steve.png', slim: true },
  { skinUrl: '/profile/player/slim/sunny.png', slim: true },
  { skinUrl: '/profile/player/slim/zuri.png', slim: true },
  { skinUrl: '/profile/player/wide/alex.png' },
  { skinUrl: '/profile/player/wide/ari.png' },
  { skinUrl: '/profile/player/wide/efe.png' },
  { skinUrl: '/profile/player/wide/kai.png' },
  { skinUrl: '/profile/player/wide/makena.png' },
  { skinUrl: '/profile/player/wide/noor.png' },
  { skinUrl: '/profile/player/wide/steve.png' },
  { skinUrl: '/profile/player/wide/sunny.png' },
  { skinUrl: '/profile/player/wide/zuri.png' },
]

export function getUUIDHashCode(uuid: string): number {
  const uuidNoDash = uuid.replace(/-/g, '')
  const mostMost = parseInt(uuidNoDash.substring(0, 8), 16)
  const mostLeast = parseInt(uuidNoDash.substring(8, 16), 16)
  const leastMost = parseInt(uuidNoDash.substring(16, 24), 16)
  const leastLeast = parseInt(uuidNoDash.substring(24, 32), 16)
  return (mostMost ^ mostLeast ^ leastMost ^ leastLeast) >>> 0
}
