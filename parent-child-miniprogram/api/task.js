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

export function listTaskClaims(familyId) {
  return request({
    url: '/api/task/claims',
    data: { familyId }
  })
}

export function claimTask(data) {
  return request({
    url: '/api/task/claim',
    method: 'POST',
    data
  })
}

export function startTaskClaim(data) {
  return request({
    url: '/api/task/startClaim',
    method: 'POST',
    data
  })
}

export function stopTaskClaim(data) {
  return request({
    url: '/api/task/stopClaim',
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

export function createTask(data) {
  return request({
    url: '/api/task/create',
    method: 'POST',
    data
  })
}

export function updateTask(data) {
  return request({
    url: '/api/task/update',
    method: 'POST',
    data
  })
}

export function archiveTask(data) {
  return request({
    url: '/api/task/archive',
    method: 'POST',
    data
  })
}
