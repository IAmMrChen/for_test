import { request } from '../utils/request.js'

export function createInvite(data) {
  return request({
    url: '/api/invite/create',
    method: 'POST',
    data
  })
}

export function acceptInvite(data) {
  return request({
    url: '/api/invite/accept',
    method: 'POST',
    data
  })
}
