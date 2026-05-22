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

export function createReward(data) {
  return request({
    url: '/api/reward/create',
    method: 'POST',
    data
  })
}

export function updateReward(data) {
  return request({
    url: '/api/reward/update',
    method: 'POST',
    data
  })
}

export function offShelfReward(data) {
  return request({
    url: '/api/reward/offShelf',
    method: 'POST',
    data
  })
}
