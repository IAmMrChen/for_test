import { request } from '../utils/request.js'

export function listMembers(familyId) {
  return request({
    url: '/api/member/list',
    data: { familyId }
  })
}

export function createVirtualChild(data) {
  return request({
    url: '/api/member/createVirtualChild',
    method: 'POST',
    data
  })
}

export function createVirtualChildBindInvite(data) {
  return request({
    url: '/api/member/createVirtualChildBindInvite',
    method: 'POST',
    data
  })
}

export function acceptVirtualChildBindInvite(data) {
  return request({
    url: '/api/member/acceptVirtualChildBindInvite',
    method: 'POST',
    data
  })
}
