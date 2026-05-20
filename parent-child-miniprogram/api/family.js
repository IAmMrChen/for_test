import { request } from '../utils/request.js'

export function listFamilies() {
  return request({ url: '/api/family/list' })
}

export function createFamily(data) {
  return request({
    url: '/api/family/create',
    method: 'POST',
    data
  })
}
