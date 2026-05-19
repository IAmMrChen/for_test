import { API_BASE_URL } from '../config/env.js'
import { getToken } from './storage.js'

export function request(options) {
  const { url, method = 'GET', data = {}, auth = true } = options
  const header = {
    ...(options.header || {})
  }
  const token = getToken()
  if (auth && token) {
    header.Authorization = `Bearer ${token}`
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url: `${API_BASE_URL}${url}`,
      method,
      data,
      header,
      success(res) {
        if (res.statusCode === 401) {
          const error = new Error('登录已过期')
          error.statusCode = 401
          reject(error)
          return
        }
        const body = res.data || {}
        if (res.statusCode >= 200 && res.statusCode < 300 && body.code === 0) {
          resolve(body.data)
          return
        }
        reject(new Error(body.message || '请求失败'))
      },
      fail() {
        reject(new Error('网络请求失败'))
      }
    })
  }).catch((error) => {
    if (error.statusCode !== 401) {
      uni.showToast({
        title: error.message || '请求失败',
        icon: 'none'
      })
    }
    throw error
  })
}
