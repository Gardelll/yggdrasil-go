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

'use client'

import React from 'react'
import { OAK_PLANKS_URL } from '@/lib/oak-planks'
import { useServerMeta, type ServerMeta } from '@/lib/server-meta'

// Steve face 8x8 pixel grid (extracted from wide/steve.png)
const STEVE_FACE = [
  ['#332411','#332411','#3F2A15','#3F2A15','#3F2A15','#3F2A15','#332411','#2B1E0D'],
  ['#241808','#332411','#332411','#3F2A15','#3F2A15','#332411','#3F2A15','#332411'],
  ['#2B1E0D','#9B6349','#B3795E','#B7836B','#B3795E','#AA7259','#9B6349','#342512'],
  ['#9B6349','#AA7259','#B3795E','#B3795E','#AA7259','#AA7259','#AA7259','#9B6349'],
  ['#AA7259','#FFFFFF','#523D89','#AA7259','#9B6349','#523D89','#FFFFFF','#AA7259'],
  ['#9B6349','#AA7259','#AA7259','#6A4030','#6A4030','#AA7259','#AA7259','#9B6349'],
  ['#90593F','#8F5E3E','#492510','#774235','#774235','#421D0A','#8F5E3E','#815339'],
  ['#94603E','#815339','#421D0A','#492510','#421D0A','#492510','#815339','#8F5E3E'],
]

// Alex face 8x8 pixel grid (extracted from slim/alex.png)
const ALEX_FACE = [
  ['#E58D3F','#F3A858','#EB983F','#F3A858','#E58D3F','#EB983F','#F3A858','#E58D3F'],
  ['#EB983F','#F3A858','#E58D3F','#EB983F','#E58D3F','#E58D3F','#EB983F','#F3A858'],
  ['#EB983F','#E58D3F','#EB983F','#E58D3F','#DFC4A2','#DFC4A2','#E58D3F','#EB983F'],
  ['#EB983F','#EB983F','#E58D3F','#EBD0B0','#EFDABF','#EBD0B0','#EBD0B0','#E58D3F'],
  ['#EBD0B0','#FFFFFF','#236224','#EFDABF','#EFDABF','#236224','#FFFFFF','#EBD0B0'],
  ['#EFDABF','#EFDABF','#EFDABF','#EFDABF','#EFDABF','#EFDABF','#EFDABF','#EFDABF'],
  ['#EFDABF','#EFDABF','#EFDABF','#EFBBB1','#EFBBB1','#EFDABF','#EFDABF','#EBD0B0'],
  ['#EBD0B0','#EFDABF','#EFDABF','#EFDABF','#EFDABF','#EFDABF','#EBD0B0','#DFC4A2'],
]

/** Renders a Minecraft 8x8 pixel face as crisp SVG */
function PixelHead({ grid, size }: { grid: string[][]; size: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 8 8"
      xmlns="http://www.w3.org/2000/svg"
      style={{ imageRendering: 'pixelated' }}
    >
      {grid.map((row, y) =>
        row.map((color, x) => (
          <rect key={`${x}-${y}`} x={x} y={y} width={1} height={1} fill={color} />
        ))
      )}
    </svg>
  )
}

/** Shared background layers (pixel pattern + glow) */
function DecorBgLayers() {
  return (
    <>
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          backgroundImage: OAK_PLANKS_URL,
          backgroundSize: '128px 128px',
          imageRendering: 'pixelated',
          opacity: 0.7,
        }}
      />
      <div
        className="absolute pointer-events-none"
        style={{
          top: '35%', left: '50%', transform: 'translate(-50%, -50%)',
          width: 280, height: 280, borderRadius: '50%',
          background: 'radial-gradient(circle, #d4a04a 0%, transparent 70%)',
          animation: 'auth-glow-pulse 6s ease-in-out infinite',
        }}
      />
    </>
  )
}

function TitleBlock({ meta, mobile }: { meta: ServerMeta | null; mobile?: boolean }) {
  const serverName = meta?.meta?.serverName || 'Yggdrasil'
  const implName = meta?.meta?.implementationName
  const implVersion = meta?.meta?.implementationVersion
  const subtitle = implName
    ? `${implName}${implVersion ? ' ' + implVersion : ''}`
    : 'Minecraft Authentication'

  if (mobile) {
    return (
      <div className="relative z-10 text-center">
        <h2
          className="text-xl font-bold"
          style={{ fontFamily: 'var(--font-cinzel), serif', color: 'rgba(255,255,255,0.88)', letterSpacing: '0.08em' }}
        >
          {serverName}
        </h2>
        <p
          className="mt-1 text-[10px] tracking-[0.25em] uppercase"
          style={{ color: 'rgba(255, 255, 255, 0.50)' }}
        >
          {subtitle}
        </p>
      </div>
    )
  }

  return (
    <div
      className="absolute bottom-0 left-0 right-0 px-10 pb-10 z-10"
      style={{ animation: 'auth-fade-up 0.8s ease-out both', animationDelay: '0.2s' }}
    >
      <h2
        className="text-[1.7rem] font-bold tracking-wide"
        style={{ fontFamily: 'var(--font-cinzel), serif', color: 'rgba(255,255,255,0.88)', letterSpacing: '0.08em' }}
      >
        {serverName}
      </h2>
      <p
        className="mt-2 text-xs tracking-[0.25em] uppercase"
        style={{ color: 'rgba(255, 255, 255, 0.50)' }}
      >
        {subtitle}
      </p>
      <div
        className="mt-4 h-px w-16"
        style={{ background: 'linear-gradient(90deg, rgba(212, 181, 122, 0.5), transparent)' }}
      />
    </div>
  )
}

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  const meta = useServerMeta()

  return (
    <div className="min-h-[calc(100vh-2rem)] flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-sm lg:max-w-4xl mx-auto rounded-2xl overflow-hidden shadow-lg border border-border">

        {/* ===== Mobile banner (lg:hidden) ===== */}
        <div
          className="lg:hidden relative overflow-hidden px-6 py-8"
          style={{ background: '#8E6B39' }}
        >
          <DecorBgLayers />
          {/* Bouncing heads */}
          <div className="relative z-10 flex items-end gap-4 mb-5 justify-center">
            <div style={{ animation: 'auth-bounce 2s ease-in-out infinite' }}>
              <PixelHead grid={STEVE_FACE} size={40} />
            </div>
            <div style={{ animation: 'auth-bounce 2s ease-in-out infinite', animationDelay: '-0.4s' }}>
              <PixelHead grid={ALEX_FACE} size={40} />
            </div>
          </div>
          <TitleBlock meta={meta} mobile />
          <div
            className="absolute inset-0 pointer-events-none"
            style={{ boxShadow: 'inset 0 0 60px rgba(0,0,0,0.4)' }}
          />
        </div>

        {/* ===== Desktop split ===== */}
        <div className="lg:grid lg:grid-cols-2 lg:items-stretch">

          {/* Left: Decorative panel (desktop only) */}
          <div
            className="hidden lg:flex relative overflow-hidden flex-col items-center justify-center"
            style={{ background: '#1a2e1a', minHeight: 480 }}
          >
            <DecorBgLayers />

            {/* Bouncing Steve & Alex heads */}
            <div className="relative z-10 flex items-end gap-6">
              <div
                className="drop-shadow-[0_8px_24px_rgba(212,181,122,0.2)]"
                style={{ animation: 'auth-bounce 2s ease-in-out infinite' }}
              >
                <PixelHead grid={STEVE_FACE} size={96} />
              </div>
              <div
                className="drop-shadow-[0_8px_24px_rgba(212,181,122,0.2)]"
                style={{ animation: 'auth-bounce 2s ease-in-out infinite', animationDelay: '-0.4s' }}
              >
                <PixelHead grid={ALEX_FACE} size={96} />
              </div>
            </div>

            <TitleBlock meta={meta} />

            {/* Vignette */}
            <div
              className="absolute inset-0 pointer-events-none"
              style={{ boxShadow: 'inset 0 0 80px rgba(0,0,0,0.4)' }}
            />
          </div>

          {/* Right: Form */}
          <div className="bg-card flex items-center justify-center px-6 py-8 lg:px-10 lg:py-12">
            <div className="w-full max-w-sm">
              {children}
            </div>
          </div>

        </div>
      </div>
    </div>
  )
}
