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
// Card wrapper provided by AuthLayout
import { Eye, EyeOff, Loader2 } from 'lucide-react'
import { SubmitHandler, useForm } from 'react-hook-form'
import axios from 'axios'
import { toast } from 'sonner'
import { useRouter } from 'next/navigation'

type Inputs = {
  username: string
  password: string
}

function PasswordReset() {
  const router = useRouter()
  const { register, handleSubmit, formState: { errors } } = useForm<Inputs>()
  const [submitting, setSubmitting] = React.useState(false)

  const onSubmit: SubmitHandler<Inputs> = data => {
    setSubmitting(true)
    const hash = window.location.hash
    if (!hash) {
      setSubmitting(false)
      toast.error('链接失效，请重新打开')
      return
    }
    const params = new URLSearchParams(hash.substring(1))
    axios.post('/authserver/resetPassword', {
      email: data.username,
      password: data.password,
      accessToken: params.get('passwordResetToken'),
    })
      .then(() => {
        toast.success('重置成功')
        router.replace('/login/')
      })
      .catch(e => {
        const response = e.response
        if (response && response.status >= 400 && response.status < 500) {
          let errorMessage = response.data.errorMessage ?? response.data
          toast.error('重置失败: ' + errorMessage)
        } else {
          toast.error('网络错误:' + e.message)
        }
      })
      .finally(() => setSubmitting(false))
  }

  const [showPassword, setShowPassword] = React.useState(false)
  const handleClickShowPassword = () => setShowPassword(show => !show)
  const handleMouseDownPassword = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault()
  }

  const toLogin = () => router.replace('/login/')

  return (
      <div>
        <h2 className="text-xl font-semibold mb-6">重置密码</h2>
          <form autoComplete="off" onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-1">
              <Label htmlFor="username-input">邮箱</Label>
              <Input
                id="username-input"
                type="email"
                required
                className={errors.username ? 'border-destructive' : ''}
                {...register('username', { required: true })}
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="password-input">新密码</Label>
              <div className="relative">
                <Input
                  id="password-input"
                  type={showPassword ? 'text' : 'password'}
                  required
                  minLength={6}
                  className={errors.password ? 'border-destructive pr-10' : 'pr-10'}
                  {...register('password', { required: true, minLength: 6 })}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
                  onClick={handleClickShowPassword}
                  onMouseDown={handleMouseDownPassword}
                  aria-label="显示密码"
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
            </div>

            <div className="flex gap-2">
              <Button type="button" variant="outline" onClick={toLogin}>登录</Button>
              <Button type="submit" disabled={submitting}>
                {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
                重置
              </Button>
            </div>
          </form>
      </div>
  )
}

export default PasswordReset
