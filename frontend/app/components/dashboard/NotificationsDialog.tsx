import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Bell, CheckCheck } from 'lucide-react'
import { Button } from '~/components/ui/button'
import { Badge } from '~/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '~/components/ui/dialog'
import api from '~/lib/api'

type Notification = {
  id: string
  product_name: string
  type: 'PRICE_DROP' | 'RESTOCK'
  old_price?: string
  new_price?: string
  read: boolean
  created_at: string
}

export function NotificationsDialog() {
  const queryClient = useQueryClient()

  const { data: notifications } = useQuery<Notification[]>({
    queryKey: ['notifications'],
    queryFn: () => api.get('/v1/notifications').then(r => r.data.notifications),
  })

  const markAllReadMutation = useMutation({
    mutationFn: () => api.patch('/v1/notifications/read-all'),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  })

  const unreadCount = notifications?.filter(n => !n.read).length ?? 0

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="w-4 h-4" />
          {unreadCount > 0 && (
            <Badge className="absolute -top-1 -right-1 h-4 w-4 p-0 flex items-center justify-center text-[10px]">
              {unreadCount}
            </Badge>
          )}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader className="flex-row items-center justify-between pr-8">
          <DialogTitle>Notifications</DialogTitle>
          {unreadCount > 0 && (
            <Button
              variant="ghost"
              size="sm"
              className="h-7 text-xs gap-1"
              onClick={() => markAllReadMutation.mutate()}
              disabled={markAllReadMutation.isPending}
            >
              <CheckCheck className="w-3 h-3" /> Mark all read
            </Button>
          )}
        </DialogHeader>
        <div className="space-y-2 max-h-80 overflow-y-auto">
          {!notifications?.length && (
            <p className="text-sm text-muted-foreground text-center py-6">
              No notifications yet.
            </p>
          )}
          {notifications?.map(n => (
            <div
              key={n.id}
              className={`p-3 rounded-md text-sm border space-y-0.5 ${
                n.read ? 'border-border text-muted-foreground' : 'border-accent/40 bg-accent/5'
              }`}
            >
              <div className="flex items-center justify-between gap-2">
                <span className="font-medium truncate">{n.product_name}</span>
                <Badge variant="outline" className="text-[10px] shrink-0">
                  {n.type === 'PRICE_DROP' ? 'Price Drop' : 'Restock'}
                </Badge>
              </div>
              {n.type === 'PRICE_DROP' && n.old_price && n.new_price && (
                <p className="text-xs">
                  <span className="line-through text-muted-foreground">{n.old_price}</span>
                  {' → '}
                  <span className="font-semibold text-accent">{n.new_price}</span>
                </p>
              )}
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  )
}
