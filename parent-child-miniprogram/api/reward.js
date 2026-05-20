import { request } from '../utils/request.js'

export function listRewards(familyId) {
  return request({
    url: '/api/reward/list',
    data: { familyId }
  })
}

export function listRewardRecords(params) {
  return request({
    url: '/api/reward/records',
    data: params
  })
}

export function applyReward(data) {
  return request({
    url: '/api/reward/apply',
    method: 'POST',
    data
  })
}

export function deliverReward(data) {
  return request({
    url: '/api/reward/deliver',
    method: 'POST',
    data
  })
}

export function rejectReward(data) {
  return request({
    url: '/api/reward/reject',
    method: 'POST',
    data
  })
}

export function receiveReward(data) {
  return request({
    url: '/api/reward/receive',
    method: 'POST',
    data
  })
}
