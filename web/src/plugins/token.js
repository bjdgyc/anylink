const tokenKey = 'AnyLink-Jwt-Token'
const tokenUser = 'AnyLink-Jwt-User'

export function getToken() {
    return localStorage.getItem(tokenKey)
}

export function setToken(token) {
    return localStorage.setItem(tokenKey, token)
}

export function setUser(username) {
    return localStorage.setItem(tokenUser, username)
}

export function getUser() {
    return localStorage.getItem(tokenUser)
}

export function removeToken() {
    return localStorage.removeItem(tokenKey)
}
