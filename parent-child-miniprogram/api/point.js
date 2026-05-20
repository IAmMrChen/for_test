import { request } from '../utils/request.js'

export function listPointLogs(params) {
  return request({
    url: '/api/point/logs',
    data: params
  })
}
