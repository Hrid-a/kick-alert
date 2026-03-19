import { Outlet, redirect } from 'react-router'
import { getToken } from '~/lib/auth'

export function clientLoader() {
  // Already authenticated — send to app
  if (getToken()) throw redirect('/dashboard')
  return null
}

export default function AuthLayout() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <Outlet />
    </div>
  )
}
