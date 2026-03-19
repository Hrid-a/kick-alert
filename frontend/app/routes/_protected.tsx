import { Outlet, redirect, useLoaderData } from 'react-router'
import { useEffect } from 'react'
import { getToken, setToken, getStoredUser } from '~/lib/auth'
import { useAuth } from '~/context/auth'
import api from '~/lib/api'

// Runs on every navigation into a protected route.
// If the in-memory token is gone (e.g. after a page reload) we try to refresh
// using the httpOnly refresh_token cookie set by the Go backend.
export async function clientLoader() {
  if (getToken()) return null

  try {
    const { data } = await api.get<{ access_token: string }>('/v1/refresh')
    setToken(data.access_token)
    return null
  } catch {
    throw redirect('/login')
  }
}

export default function ProtectedLayout() {
  useLoaderData<typeof clientLoader>()
  const { user, setAuth } = useAuth()

  // After a page reload the context is empty but the token was just re-hydrated
  // by the loader and the user profile is in localStorage — sync them.
  useEffect(() => {
    if (!user) {
      const token = getToken()
      const stored = getStoredUser()
      if (token && stored) setAuth(token, stored)
    }
  }, [user, setAuth])

  return <Outlet />
}
