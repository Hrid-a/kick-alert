// Module-level token store — kept in memory so it never touches localStorage.
// Used by the axios interceptor which cannot call React hooks.

export type AuthUser = {
  id: string
  name: string
  email: string
}

let _token: string | null = null

export const getToken = () => _token
export const setToken = (t: string | null) => { _token = t }

// Non-sensitive user info (name, email) is stored in localStorage so it
// survives page reloads. The access token itself stays in memory only.
export const getStoredUser = (): AuthUser | null => {
  try {
    const raw = localStorage.getItem('kick_user')
    return raw ? (JSON.parse(raw) as AuthUser) : null
  } catch {
    return null
  }
}

export const storeUser = (user: AuthUser) => {
  localStorage.setItem('kick_user', JSON.stringify(user))
}

export const clearStoredUser = () => {
  localStorage.removeItem('kick_user')
}
