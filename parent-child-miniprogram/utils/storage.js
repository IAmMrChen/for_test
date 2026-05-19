const TOKEN_KEY = 'authToken'
const CURRENT_USER_KEY = 'currentUser'
const CURRENT_FAMILY_KEY = 'currentFamily'

export function getToken() {
  return uni.getStorageSync(TOKEN_KEY) || ''
}

export function setToken(token) {
  uni.setStorageSync(TOKEN_KEY, token)
}

export function clearToken() {
  uni.removeStorageSync(TOKEN_KEY)
}

export function getCurrentUser() {
  return uni.getStorageSync(CURRENT_USER_KEY) || null
}

export function setCurrentUser(user) {
  uni.setStorageSync(CURRENT_USER_KEY, user)
}

export function getCurrentFamily() {
  return uni.getStorageSync(CURRENT_FAMILY_KEY) || null
}

export function setCurrentFamily(family) {
  uni.setStorageSync(CURRENT_FAMILY_KEY, family)
}

export function clearCurrentFamily() {
  uni.removeStorageSync(CURRENT_FAMILY_KEY)
}
