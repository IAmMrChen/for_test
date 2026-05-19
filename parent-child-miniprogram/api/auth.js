import { DEMO_USER } from '../config/env.js'
import { request } from '../utils/request.js'
import { clearToken, getToken, setCurrentUser, setToken } from '../utils/storage.js'

export async function demoLogin(payload = DEMO_USER) {
  const result = await request({
    url: '/api/auth/demoLogin',
    method: 'POST',
    data: payload,
    auth: false
  })
  setToken(result.token)
  setCurrentUser(result.user)
  return result
}

export async function ensureDemoLogin(force = false) {
  if (!force && getToken()) {
    return
  }
  if (force) {
    clearToken()
  }
  await demoLogin()
}
