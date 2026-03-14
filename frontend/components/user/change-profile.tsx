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
import { Loader2 } from 'lucide-react'
import { AppState } from '@/lib/types'
import { SubmitHandler, useForm } from 'react-hook-form'
import axios from 'axios'
import { toast } from 'sonner'
import { loadAuthState, saveAuthState } from '@/lib/auth'

type ChangeProfileInputs = {
  changeTo: string
}

export default function ChangeProfileForm(props: {
  appData: AppState
  onProfileNameChange: (name: string) => void
}) {
  const { appData, onProfileNameChange } = props
  const [submitting, setSubmitting] = React.useState(false)
  const { register, handleSubmit, formState: { errors } } = useForm<ChangeProfileInputs>()

  const onSubmit: SubmitHandler<ChangeProfileInputs> = data => {
    setSubmitting(true)
    axios.post('/authserver/change', {
      accessToken: appData.accessToken,
      changeTo: data.changeTo,
    }).then(() => {
      toast.success('更改成功')
      const updated = { ...loadAuthState(), profileName: data.changeTo }
      saveAuthState(updated)
      onProfileNameChange(data.changeTo)
    }).catch(e => {
      const response = e.response
      if (response && response.data) {
        let errorMessage = response.data.errorMessage
        let message = '更改失败: ' + errorMessage
        if (errorMessage === 'profileName exist') {
          message = '更改失败: 角色名已存在'
        } else if (errorMessage === 'profileName duplicate') {
          message = '更改失败: 角色名与正版用户冲突'
        }
        toast.error(message)
      } else {
        toast.error('网络错误:' + e.message)
      }
    }).finally(() => setSubmitting(false))
  }

  return (
    <>
      <h3 className="text-base font-semibold mb-4">更改角色名</h3>
      <form autoComplete="off" onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="space-y-1">
          <Label htmlFor="changeTo-input">角色名</Label>
          <Input
            id="changeTo-input"
            required
            defaultValue={appData.profileName}
            minLength={2}
            maxLength={16}
            className={errors.changeTo ? 'border-destructive' : ''}
            {...register('changeTo', {
              required: true,
              minLength: 2,
              pattern: /^[a-zA-Z0-9_]{1,16}$/,
              maxLength: 16,
            })}
          />
          <p className="text-xs text-muted-foreground">字母，数字或下划线</p>
        </div>
        <div>
          <Button type="submit" disabled={submitting}>
            {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
            更改
          </Button>
        </div>
      </form>
    </>
  )
}
