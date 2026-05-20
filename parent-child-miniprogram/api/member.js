import { request } from '../utils/request.js'

export function listMembers(familyId) {
  return request({
    url: '/api/member/list',
    data: { familyId }
  })
}
