import { request } from '../utils/request.js'

export function listTasks(familyId) {
  return request({
    url: '/api/task/list',
    data: { familyId }
  })
}

export function listTaskRecords(params) {
  return request({
    url: '/api/task/records',
    data: params
  })
}

export function claimTask(data) {
  return request({
    url: '/api/task/claim',
    method: 'POST',
    data
  })
}

export function submitTask(data) {
  return request({
    url: '/api/task/submit',
    method: 'POST',
    data
  })
}

export function auditTask(data) {
  return request({
    url: '/api/task/audit',
    method: 'POST',
    data
  })
}
