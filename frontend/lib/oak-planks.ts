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

// Oak plank 16x16 tile (extracted from zh.minecraft.wiki/images/BlockSprite_oak-planks.png)
const OAK_PLANKS: string[][] = [
  ['#B28C56','#A9874C','#B28C56','#BD9659','#BD9659','#BD9659','#BD9659','#BD9659','#B28C56','#BD9659','#BD9659','#BD9659','#BD9659','#BD9659','#B28C56','#8E6B39'],
  ['#B28C56','#B28C56','#A9874C','#A9874C','#8E6B39','#987C44','#A9874C','#B28C56','#B28C56','#B28C56','#A9874C','#A9874C','#B28C56','#B28C56','#987C44','#987C44'],
  ['#A9874C','#B28C56','#B28C56','#B28C56','#A9874C','#B28C56','#A9874C','#987C44','#987C44','#987C44','#987C44','#A9874C','#A9874C','#B28C56','#B28C56','#8E6B39'],
  ['#8E6B39','#75592F','#75592F','#8E6B39','#8E6B39','#75592F','#5E4725','#5E4725','#75592F','#75592F','#5E4725','#75592F','#5E4725','#5E4725','#75592F','#5E4725'],
  ['#B28C56','#BD9659','#987C44','#BD9659','#BD9659','#BD9659','#BD9659','#987C44','#BD9659','#BD9659','#BD9659','#B28C56','#A9874C','#A9874C','#987C44','#B28C56'],
  ['#A9874C','#B28C56','#B28C56','#A9874C','#987C44','#A9874C','#987C44','#8E6B39','#987C44','#A9874C','#A9874C','#B28C56','#B28C56','#B28C56','#B28C56','#A9874C'],
  ['#987C44','#987C44','#A9874C','#B28C56','#A9874C','#A9874C','#A9874C','#8E6B39','#B28C56','#B28C56','#B28C56','#A9874C','#A9874C','#987C44','#987C44','#987C44'],
  ['#5E4725','#5E4725','#75592F','#75592F','#8E6B39','#75592F','#5E4725','#5E4725','#5E4725','#5E4725','#75592F','#75592F','#75592F','#8E6B39','#75592F','#5E4725'],
  ['#B28C56','#BD9659','#BD9659','#B28C56','#A9874C','#A9874C','#BD9659','#BD9659','#BD9659','#BD9659','#BD9659','#BD9659','#BD9659','#BD9659','#B28C56','#987C44'],
  ['#B28C56','#A9874C','#B28C56','#B28C56','#B28C56','#B28C56','#A9874C','#987C44','#987C44','#987C44','#A9874C','#987C44','#A9874C','#987C44','#987C44','#8E6B39'],
  ['#BD9659','#B28C56','#A9874C','#A9874C','#987C44','#987C44','#987C44','#987C44','#A9874C','#A9874C','#A9874C','#A9874C','#987C44','#987C44','#B28C56','#987C44'],
  ['#5E4725','#5E4725','#75592F','#8E6B39','#8E6B39','#75592F','#5E4725','#5E4725','#5E4725','#75592F','#8E6B39','#75592F','#75592F','#5E4725','#5E4725','#5E4725'],
  ['#BD9659','#987C44','#B28C56','#BD9659','#BD9659','#B28C56','#B28C56','#987C44','#BD9659','#BD9659','#BD9659','#987C44','#BD9659','#B28C56','#BD9659','#BD9659'],
  ['#A9874C','#A9874C','#B28C56','#B28C56','#987C44','#987C44','#A9874C','#987C44','#987C44','#B28C56','#B28C56','#A9874C','#B28C56','#A9874C','#A9874C','#A9874C'],
  ['#A9874C','#987C44','#987C44','#A9874C','#B28C56','#A9874C','#987C44','#8E6B39','#B28C56','#B28C56','#A9874C','#987C44','#987C44','#987C44','#987C44','#987C44'],
  ['#5E4725','#75592F','#75592F','#5E4725','#5E4725','#75592F','#8E6B39','#8E6B39','#8E6B39','#75592F','#5E4725','#75592F','#5E4725','#5E4725','#5E4725','#5E4725'],
]

function buildUrl(): string {
  const W = 16
  let rects = ''
  for (let y = 0; y < W; y++) {
    for (let x = 0; x < W; x++) {
      rects += `<rect x="${x}" y="${y}" width="1" height="1" fill="${OAK_PLANKS[y][x]}"/>`
    }
  }
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${W}" viewBox="0 0 ${W} ${W}" shape-rendering="crispEdges">${rects}</svg>`
  return `url("data:image/svg+xml,${encodeURIComponent(svg)}")`
}

export const OAK_PLANKS_URL = buildUrl()
