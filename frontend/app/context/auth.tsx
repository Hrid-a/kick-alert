import { createContext, useContext, useState, useCallback } from 'react'
import { setToken, storeUser, clearStoredUser, getToken, getStoredUser } from '~/lib/auth'
import type { AuthUser } from '~/lib/auth'

type AuthContextType = {
  user: AuthUser | null
  setAuth: (token: string, user: AuthUser) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  // Seed initial state from localStorage so a page reload doesn't flash the login screen
  const [user, setUser] = useState<AuthUser | null>(() => {
    // Token is in-memory so it's gone after reload — only use stored user
    // if the protected loader has already re-hydrated the token.
    return getToken() ? getStoredUser() : null
  })

  const setAuth = useCallback((token: string, userData: AuthUser) => {
    setToken(token)
    storeUser(userData)
    setUser(userData)
  }, [])

  const logout = useCallback(() => {
    setToken(null)
    clearStoredUser()
    setUser(null)
  }, [])

  return (
    <AuthContext.Provider value={{ user, setAuth, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
