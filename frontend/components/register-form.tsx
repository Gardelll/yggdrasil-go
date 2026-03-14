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
  profileName: string
  password: string
}

export default function RegisterForm() {
  const router = useRouter()
  const { register, handleSubmit, formState: { errors } } = useForm<Inputs>()
  const [submitting, setSubmitting] = React.useState(false)
  const [showPassword, setShowPassword] = React.useState(false)

  const onSubmit: SubmitHandler<Inputs> = data => {
    setSubmitting(true)
    axios.post('/authserver/register', {
      username: data.username,
      password: data.password,
      profileName: data.profileName,
    })
      .then(response => {
        const d = response.data
        if (d && d.id) {
          toast.success('注册成功')
          router.replace('/login/')
        } else {
          toast.error(d && d.errorMessage ? '注册失败: ' + d.errorMessage : '注册失败')
        }
      })
      .catch(e => {
        const response = e.response
        if (response && response.data) {
          let errorMessage = typeof response.data.errorMessage === 'string'
            ? response.data.errorMessage
            : typeof response.data === 'string' ? '服务器错误' : '注册失败'
          let message = '注册失败: ' + errorMessage
          if (errorMessage === 'profileName exist') {
            message = '注册失败: 角色名已存在'
          } else if (errorMessage === 'profileName duplicate') {
            message = '注册失败: 角色名与正版用户冲突'
          }
          toast.error(message)
        } else {
          toast.error('网络错误:' + e.message)
        }
      })
      .finally(() => setSubmitting(false))
  }

  return (
      <div>
        <h2 className="text-xl font-semibold mb-6">注册</h2>
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
              <Label htmlFor="profileName-input">角色名</Label>
              <Input
                id="profileName-input"
                required
                minLength={2}
                maxLength={16}
                className={errors.profileName ? 'border-destructive' : ''}
                {...register('profileName', {
                  required: true,
                  minLength: 2,
                  pattern: /^[a-zA-Z0-9_]{1,16}$/,
                  maxLength: 16,
                })}
              />
              <p className="text-xs text-muted-foreground">字母，数字或下划线</p>
            </div>

            <div className="space-y-1">
              <Label htmlFor="password-input">密码</Label>
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
                  onClick={() => setShowPassword(s => !s)}
                  onMouseDown={e => e.preventDefault()}
                  aria-label="显示密码"
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">请妥善保管密码</p>
            </div>

            <div className="flex flex-wrap gap-2">
              <Button type="button" variant="outline" onClick={() => router.push('/login/')}>已有账号登录</Button>
              <Button type="submit" disabled={submitting}>
                {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
                注册
              </Button>
            </div>
          </form>
      </div>
  )
}
