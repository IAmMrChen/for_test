import { request } from '../utils/request.js'

export function getDashboardSummary(familyId) {
  return request({
    url: '/api/dashboard/summary',
    data: { familyId }
  })
}
