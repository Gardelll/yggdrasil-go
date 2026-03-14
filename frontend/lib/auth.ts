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

import { AppState } from '@/lib/types'

const STORAGE_KEY = 'appData'

export const defaultAppState: AppState = {
  accessToken: '',
  tokenValid: false,
  loginTime: 0,
  profileName: '',
  uuid: '',
}

export function loadAuthState(): AppState {
  if (typeof window === 'undefined') return { ...defaultAppState }
  const saved = localStorage.getItem(STORAGE_KEY)
  if (!saved) return { ...defaultAppState }
  try {
    return { ...defaultAppState, ...JSON.parse(saved) }
  } catch {
    return { ...defaultAppState }
  }
}

export function saveAuthState(s: AppState) {
  if (typeof window === 'undefined') return
  localStorage.setItem(STORAGE_KEY, JSON.stringify(s))
}

export function clearAuthState() {
  saveAuthState({ ...defaultAppState })
}

export function isTokenFresh(s: AppState): boolean {
  return !!(s.tokenValid && s.accessToken && Date.now() - s.loginTime < 30 * 86400 * 1000)
}
