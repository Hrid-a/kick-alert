import { useNavigate } from 'react-router'
import { useAuth } from '~/context/auth'
import { DashboardNavbar } from '~/components/dashboard/DashboardNavbar'
import { TrackForm } from '~/components/dashboard/TrackForm'
import { WatchlistList } from '~/components/dashboard/WatchlistList'
import api from '~/lib/api'

export default function DashboardPage() {
  const navigate = useNavigate()
  const { user, logout } = useAuth()

  const handleLogout = async () => {
    try { await api.post('/v1/logout') } catch {}
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="min-h-screen bg-background">
      <DashboardNavbar user={user} onLogout={handleLogout} />

      <main className="max-w-5xl mx-auto px-6 py-10 space-y-8">
        <div>
          <h1 className="text-2xl font-display font-bold uppercase tracking-tight">
            Welcome back{user?.name ? `, ${user.name.split(' ')[0]}` : ''}
          </h1>
          <p className="text-muted-foreground text-sm mt-1">
            Track your Kick products and get alerted on price drops.
          </p>
        </div>

        <TrackForm />

        <section className="space-y-3">
          <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
            Tracked Products
          </h2>
          <WatchlistList />
        </section>
      </main>
    </div>
  )
}
