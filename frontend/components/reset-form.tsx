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
import { Loader2 } from 'lucide-react'
import { SubmitHandler, useForm } from 'react-hook-form'
import axios from 'axios'
import { toast } from 'sonner'
import { useRouter } from 'next/navigation'

type Inputs = {
  username: string
}

export default function ResetForm() {
  const router = useRouter()
  const { register, handleSubmit, formState: { errors } } = useForm<Inputs>()
  const [submitting, setSubmitting] = React.useState(false)
  const [countdown, setCountdown] = React.useState(0)
  const timerRef = React.useRef<ReturnType<typeof setInterval>>(undefined)

  React.useEffect(() => {
    return () => { clearInterval(timerRef.current) }
  }, [])

  const onSubmit: SubmitHandler<Inputs> = data => {
    setSubmitting(true)
    setCountdown(60)

    clearInterval(timerRef.current)
    timerRef.current = setInterval(() => {
      setCountdown(prev => {
        if (prev <= 1) {
          clearInterval(timerRef.current)
          setSubmitting(false)
          return 0
        }
        return prev - 1
      })
    }, 1000)

    axios.post('/authserver/sendEmail', {
      email: data.username,
      emailType: 'resetPassword',
    })
      .then(() => {
        toast.success('重置链接发送成功，请检查垃圾邮箱')
      })
      .catch(e => {
        const response = e.response
        if (response && response.data) {
          let errorMessage = typeof response.data.errorMessage === 'string'
            ? response.data.errorMessage
            : typeof response.data === 'string' ? '服务器错误' : '发送失败'
          toast.error('发送失败: ' + errorMessage)
        } else {
          toast.error('网络错误:' + e.message)
        }
        clearInterval(timerRef.current)
        setSubmitting(false)
        setCountdown(0)
      })
  }

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

            <div className="flex flex-wrap gap-2">
              <Button type="button" variant="outline" onClick={() => router.push('/login/')}>已有账号登录</Button>
              <Button type="button" variant="outline" onClick={() => router.push('/register/')}>注册</Button>
              <Button type="submit" disabled={submitting}>
                {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
                {submitting ? `${countdown}` : '重置'}
              </Button>
            </div>
          </form>
      </div>
  )
}
