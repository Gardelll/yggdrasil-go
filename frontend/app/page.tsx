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
import { Loader2 } from 'lucide-react'
import { isTokenFresh, loadAuthState } from '@/lib/auth'
import axios from 'axios'
import { toast } from 'sonner'

export default function Home() {
  const router = useRouter()

  React.useEffect(() => {
    const hash = window.location.hash
    const params = new URLSearchParams(hash.substring(1))
    const dest = isTokenFresh(loadAuthState()) ? '/user/' : '/login/'

    if (params.has('emailVerifyToken')) {
      const token = params.get('emailVerifyToken')
      axios.get('/authserver/verifyEmail', { params: { access_token: token } })
        .then(() => {
          toast.success('邮箱验证通过')
        })
        .catch(e => {
          const response = e.response
          if (response && response.status >= 400 && response.status < 500) {
            let errorMessage = response.data.errorMessage ?? response.data
            toast.error('邮箱验证失败: ' + errorMessage)
          } else {
            toast.error('网络错误:' + e.message)
          }
        })
        .finally(() => {
          router.replace(dest)
        })
    } else {
      router.replace(dest)
    }
  }, [router])

  return (
    <div className="flex justify-center items-center min-h-[50vh]">
      <Loader2 className="animate-spin h-8 w-8" />
    </div>
  )
}
