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
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Trash2, Loader2 } from 'lucide-react'
import { AppState, SkinData } from '@/lib/types'
import axios from 'axios'
import { toast } from 'sonner'
import { fetchSkinData, getDefaultSkin } from '@/lib/profile'

function handleMouseDown(event: React.MouseEvent<HTMLButtonElement>) {
  event.preventDefault()
}

export default function UploadTextureForm(props: {
  appData: AppState
  skinData: SkinData | null
  setSkinData: React.Dispatch<React.SetStateAction<SkinData | null>>
}) {
  const { appData, skinData, setSkinData } = props
  const [submitting, setSubmitting] = React.useState(false)

  const fileInputElem = React.useRef<HTMLInputElement>(null)
  const [filePath, setFilePath] = React.useState('')
  const [url, setUrl] = React.useState('')
  const [type, setType] = React.useState('skin')
  const [model, setModel] = React.useState('default')

  const handleFilePathChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setFilePath(event.target.value)

    if (skinData) {
      if (type === 'cape' && skinData.capeUrl?.startsWith('blob:')) {
        URL.revokeObjectURL(skinData.capeUrl)
      } else if (skinData.skinUrl.startsWith('blob:')) {
        URL.revokeObjectURL(skinData.skinUrl)
      }
    }

    const fileInput = event.target
    const fileBlob = fileInput.files?.length ? fileInput.files[0] : null
    if (fileBlob && skinData) {
      const data: SkinData = { ...skinData }
      const fakeUrl = URL.createObjectURL(fileBlob)
      if (type === 'cape') {
        data.capeUrl = fakeUrl
      } else {
        data.slim = model === 'slim'
        data.skinUrl = fakeUrl
      }
      setSkinData(data)
    }
  }

  const handleModelChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setModel(event.target.value)
    if (skinData) {
      setSkinData({ ...skinData, slim: event.target.value === 'slim' })
    }
  }

  const uploadTexture = (event: React.FormEvent) => {
    event.preventDefault()
    setSubmitting(true)
    const fileInput = fileInputElem.current
    const fileBlob = fileInput?.files?.length ? fileInput.files[0] : null
    if (filePath && fileBlob) {
      const formData = new FormData()
      formData.append('model', model)
      formData.append('file', fileBlob)
      axios.put(`/api/user/profile/${appData.uuid}/${type}`, formData, {
        headers: { 'Authorization': 'Bearer ' + appData.accessToken },
      }).then(() => {
        toast.success('上传成功')
        fetchSkinData(appData.uuid).then(data => setSkinData(data))
      }).catch(e => {
        const response = e.response
        if (response && response.data) {
          toast.error(response.data.errorMessage)
        } else {
          toast.error('网络错误:' + e.message)
        }
      }).finally(() => setSubmitting(false))
    } else if (url) {
      axios.post(`/api/user/profile/${appData.uuid}/${type}`, { model, url }, {
        headers: { 'Authorization': 'Bearer ' + appData.accessToken },
      }).then(() => {
        toast.success('上传成功')
        fetchSkinData(appData.uuid).then(data => setSkinData(data))
      }).catch(e => {
        const response = e.response
        if (response && response.data) {
          toast.error(response.data.errorMessage)
        } else {
          toast.error('网络错误:' + e.message)
        }
      }).finally(() => setSubmitting(false))
    } else {
      toast.warning('未选择文件')
      setSubmitting(false)
    }
  }

  const deleteTexture = () => {
    setSubmitting(true)
    axios.delete(`/api/user/profile/${appData.uuid}/${type}`, {
      headers: { 'Authorization': 'Bearer ' + appData.accessToken },
    }).then(() => {
      toast.success('删除成功')
      if (skinData != null) {
        if (type === 'cape') {
          setSkinData({ ...skinData, capeUrl: undefined })
        } else {
          setSkinData({ ...getDefaultSkin(appData.uuid), capeUrl: skinData.capeUrl })
        }
      }
    }).catch(e => {
      const response = e.response
      if (response && response.data) {
        toast.error(response.data.errorMessage)
      } else {
        toast.error('网络错误:' + e.message)
      }
    }).finally(() => setSubmitting(false))
  }

  return (
    <>
      <h3 className="text-base font-semibold mb-4">上传材质</h3>
      <form autoComplete="off" onSubmit={uploadTexture} className="space-y-4">
        <div className="flex flex-wrap gap-6">
          <div>
            <Label className="mb-2 block">材质类别:</Label>
            <div className="flex gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="radio" name="type" value="skin" checked={type === 'skin'} onChange={e => setType(e.target.value)} />
                皮肤
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="radio" name="type" value="cape" checked={type === 'cape'} onChange={e => setType(e.target.value)} />
                披风
              </label>
            </div>
          </div>

          {type === 'skin' && (
            <div>
              <Label className="mb-2 block">材质模型:</Label>
              <div className="flex gap-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="radio" name="model" value="default" checked={model === 'default'} onChange={handleModelChange} />
                  Steve
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="radio" name="model" value="slim" checked={model === 'slim'} onChange={handleModelChange} />
                  Alex
                </label>
              </div>
            </div>
          )}
        </div>

        {!filePath && (
          <div className="space-y-1">
            <Label htmlFor="url-input">材质 URL</Label>
            <Input
              id="url-input"
              type="url"
              required={!filePath}
              value={url}
              onChange={e => setUrl(e.target.value)}
            />
          </div>
        )}

        {!url && (
          <div className="space-y-1">
            <Label htmlFor="file-input">或者选择一个图片</Label>
            <div className="relative">
              <Input
                id="file-input"
                required={!url}
                value={filePath}
                readOnly
                onClick={() => fileInputElem.current?.click()}
                className="cursor-pointer pr-10"
                placeholder="点击选择文件..."
              />
              {filePath && (
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
                  onMouseDown={handleMouseDown}
                  onClick={() => setFilePath('')}
                  aria-label="清空选择"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              )}
            </div>
            <input
              id="file-input-real"
              type="file"
              name="file"
              hidden
              ref={fileInputElem}
              accept="image/*"
              value={filePath}
              onChange={handleFilePathChange}
            />
          </div>
        )}

        <div className="flex gap-2">
          <Button type="submit" disabled={submitting}>
            {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
            上传
          </Button>
          <Button type="button" variant="outline" onClick={deleteTexture} disabled={submitting}>
            删除材质
          </Button>
        </div>
      </form>
    </>
  )
}
