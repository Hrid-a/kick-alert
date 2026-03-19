import { Link } from 'react-router'
import { Settings, LogOut } from 'lucide-react'
import { Button } from '~/components/ui/button'
import { NotificationsDialog } from './NotificationsDialog'
import type { AuthUser } from '~/lib/auth'

type Props = {
  user: AuthUser | null
  onLogout: () => void
}

export function DashboardNavbar({ user: _user, onLogout }: Props) {
  return (
    <header className="border-b border-border bg-card">
      <div className="max-w-5xl mx-auto px-6 h-14 flex items-center justify-between">
        <span className="font-display text-lg font-bold uppercase tracking-tight">KickAlert</span>

        <nav className="flex items-center gap-1">
          
          <NotificationsDialog />

          <Button asChild variant="ghost" size="icon">
            <Link to="/settings"><Settings className="w-4 h-4" /></Link>
          </Button>

          <Button variant="ghost" size="icon" onClick={onLogout}>
            <LogOut className="w-4 h-4" />
          </Button>
        </nav>
      </div>
    </header>
  )
}
