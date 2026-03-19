import { redirect } from 'react-router'
import { getToken } from '~/lib/auth'

export function clientLoader() {
  if (getToken()) throw redirect('/dashboard')
  throw redirect('/login')
}

export default function Home() {
  return null
}
