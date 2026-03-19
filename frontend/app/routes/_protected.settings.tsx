import { useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { ArrowLeft, Crown } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import { Switch } from '~/components/ui/switch'
import { Label } from '~/components/ui/label'
import { Separator } from '~/components/ui/separator'
import { Badge } from '~/components/ui/badge'
import { useAuth } from '~/context/auth'
import api from '~/lib/api'

export default function SettingsPage() {
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const [notifyEmail, setNotifyEmail] = useState(true)

  const notifyMutation = useMutation({
    mutationFn: (value: boolean) =>
      api.patch('/v1/users/me/notifications', { notify_email: value }),
    onSuccess: (_, value) => {
      setNotifyEmail(value)
      toast.success('Preferences saved.')
    },
    onError: (error: { response?: { status?: number; data?: { error?: string } } }) => {
      if (error.response?.status === 403) {
        toast.error('Notification preferences are a Pro feature.')
      } else {
        toast.error(error.response?.data?.error ?? 'Failed to save preferences.')
      }
    },
  })

  const handleLogout = async () => {
    try { await api.post('/v1/logout') } catch {}
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-2xl mx-auto px-6 py-10 space-y-6">
        {/* Back */}
        <Button asChild variant="ghost" size="sm" className="-ml-2">
          <Link to="/dashboard">
            <ArrowLeft className="w-4 h-4 mr-1" /> Back
          </Link>
        </Button>

        <div>
          <h1 className="text-2xl font-display font-bold uppercase tracking-tight">Settings</h1>
          <p className="text-muted-foreground text-sm mt-1">Manage your account and preferences.</p>
        </div>

        {/* Profile */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Profile</CardTitle>
            <CardDescription>Your account information.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Name</Label>
              <p className="text-sm font-medium">{user?.name ?? '—'}</p>
            </div>
            <Separator />
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Email</Label>
              <p className="text-sm font-medium">{user?.email ?? '—'}</p>
            </div>
          </CardContent>
        </Card>

        {/* Notifications */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <CardTitle className="text-base">Notifications</CardTitle>
              <Badge variant="outline" className="text-[10px] gap-1 h-5 px-1.5">
                <Crown className="w-2.5 h-2.5" /> Pro
              </Badge>
            </div>
            <CardDescription>
              Control how you receive price drop and restock alerts.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="notify-email">Email alerts</Label>
                <p className="text-xs text-muted-foreground">
                  Receive notifications by email when a tracked product drops in price or restocks.
                </p>
              </div>
              <Switch
                id="notify-email"
                checked={notifyEmail}
                disabled={notifyMutation.isPending}
                onCheckedChange={(checked) => notifyMutation.mutate(checked)}
              />
            </div>
          </CardContent>
        </Card>

        {/* Danger zone */}
        <Card className="border-destructive/40">
          <CardHeader>
            <CardTitle className="text-base text-destructive">Danger Zone</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <p className="text-sm font-medium">Sign out</p>
                <p className="text-xs text-muted-foreground">Sign out of your account on this device.</p>
              </div>
              <Button variant="destructive" size="sm" onClick={handleLogout}>
                Sign out
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
