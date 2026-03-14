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
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { LogOut, Loader2 } from 'lucide-react'
import { AppState, SkinData } from '@/lib/types'
import { loadAuthState, saveAuthState, clearAuthState, isTokenFresh } from '@/lib/auth'
import { fetchSkinData } from '@/lib/profile'
import { useServerMeta } from '@/lib/server-meta'
import axios from 'axios'
import { toast } from 'sonner'
import dynamic from 'next/dynamic'
import UploadTextureForm from '@/components/user/upload-texture'
import ChangeProfileForm from '@/components/user/change-profile'

const SkinRender = dynamic(() => import('@/components/skin-render/skin-render'), {
  ssr: false,
  loading: () => (
    <div className="flex justify-center items-center h-full min-h-[300px]">
      <Loader2 className="animate-spin h-8 w-8" />
    </div>
  ),
})

export default function UserPage() {
  const router = useRouter()
  useServerMeta()
  const [appState, setAppState] = React.useState<AppState | null>(null)
  const [skinData, setSkinData] = React.useState<SkinData | null>(null)

  React.useEffect(() => {
    const state = loadAuthState()
    if (!isTokenFresh(state)) {
      router.replace('/login/')
      return
    }

    const postData = { accessToken: state.accessToken }
    axios.post('/authserver/validate', postData)
      .then(() => {
        setAppState(state)
      })
      .catch(e => {
        const response = e.response
        if (response && response.status === 403) {
          axios.post('/authserver/refresh', postData)
            .then(res => {
              const d = res.data
              if (d && d.accessToken) {
                const updated: AppState = {
                  ...state,
                  accessToken: d.accessToken,
                  loginTime: Date.now(),
                  profileName: d.selectedProfile?.name ?? state.profileName,
                  uuid: d.selectedProfile?.id ?? state.uuid,
                }
                saveAuthState(updated)
                setAppState(updated)
                toast.success('已自动刷新登录状态')
              } else {
                clearAuthState()
                router.replace('/login/')
              }
            })
            .catch(re => {
              const resp = re.response
              if (resp && resp.status === 403) {
                toast.warning('登录已过期')
              } else {
                toast.error('网络错误:' + re.message)
              }
              clearAuthState()
              router.replace('/login/')
            })
        } else {
          toast.error('网络错误:' + e.message)
          router.replace('/login/')
        }
      })
  }, [router])

  React.useEffect(() => {
    if (!appState) return
    setSkinData(null)
    fetchSkinData(appState.uuid)
      .then(data => setSkinData(data))
      .catch(() => toast.error('加载皮肤失败'))
  }, [appState])

  const logout = () => {
    if (!appState) return
    axios.post('/authserver/signout', { accessToken: appState.accessToken })
      .catch(() => {})
      .finally(() => {
        clearAuthState()
        router.replace('/login/')
      })
  }

  if (!appState) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <Loader2 className="animate-spin h-8 w-8" />
      </div>
    )
  }

  return (
    <div className="min-h-[calc(100vh-2rem)] flex items-center justify-center px-4 py-8">
    <div className="w-full max-w-7xl min-[2560px]:max-w-[70vw]">
      {/* Unified card container */}
      <div className="rounded-xl border border-border overflow-hidden shadow-sm bg-card">
        {/* Header bar */}
        <div className="flex items-center justify-between px-6 min-[2560px]:px-10 py-4 min-[2560px]:py-6 border-b border-border">
          <h1 className="text-lg min-[2560px]:text-2xl font-semibold">{appState.profileName}</h1>
          <Button type="button" variant="ghost" size="sm" onClick={logout}>
            <LogOut className="h-4 w-4 mr-1.5" />
            退出
          </Button>
        </div>

        {/* Content: preview + forms side by side on desktop */}
        <div className="lg:grid lg:grid-cols-5">
          {/* Left: 3D preview */}
          <div className="lg:col-span-2 lg:border-r border-b lg:border-b-0 border-border flex items-center">
            <div className="h-[360px] lg:h-[480px] w-full">
              {skinData && (
                <SkinRender skinUrl={skinData.skinUrl} capeUrl={skinData.capeUrl} slim={skinData.slim} />
              )}
            </div>
          </div>

          {/* Right: forms */}
          <div className="lg:col-span-3 px-6 min-[2560px]:px-10 py-6 min-[2560px]:py-10 space-y-8 min-[2560px]:space-y-12">
            <UploadTextureForm appData={appState} skinData={skinData} setSkinData={setSkinData} />
            <div className="h-px bg-border" />
            <ChangeProfileForm
              appData={appState}
              onProfileNameChange={name => setAppState(s => s ? { ...s, profileName: name } : s)}
            />
          </div>
        </div>
      </div>
    </div>
    </div>
  )
}
