import { request } from '../utils/request.js'

export function listFamilies() {
  return request({ url: '/api/family/list' })
}
