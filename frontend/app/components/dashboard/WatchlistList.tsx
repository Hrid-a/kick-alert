import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { Package, ExternalLink } from 'lucide-react'
import { Card, CardContent } from '~/components/ui/card'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Skeleton } from '~/components/ui/skeleton'
import api from '~/lib/api'

type WatchlistEntry = {
  id: string
  product_id: string
  product_name: string
  product_image_url: string
  product_current_price: string
  product_currency: string
  product_in_stock: boolean
}

export function WatchlistList() {
  const { data: watchlist, isLoading } = useQuery<WatchlistEntry[]>({
    queryKey: ['watchlist'],
    queryFn: () => api.get('/v1/watchlist').then(r => r.data.watchlist),
  })

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map(i => (
          <Skeleton key={i} className="h-20 w-full rounded-lg" />
        ))}
      </div>
    )
  }

  if (!watchlist?.length) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center gap-2 py-12">
          <Package className="w-8 h-8 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">
            No products tracked yet. Paste a URL above to start.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-3">
      {watchlist.map(entry => (
        <Card key={entry.id} className="hover:border-accent/50 transition-colors">
          <CardContent className="flex items-center gap-4 py-4">
            {entry.product_image_url ? (
              <img
                src={entry.product_image_url}
                alt={entry.product_name}
                className="w-14 h-14 object-cover rounded-md shrink-0"
              />
            ) : (
              <div className="w-14 h-14 bg-muted rounded-md flex items-center justify-center shrink-0">
                <Package className="w-6 h-6 text-muted-foreground" />
              </div>
            )}

            <div className="flex-1 min-w-0">
              <p className="font-medium truncate">{entry.product_name}</p>
              <Badge
                variant={entry.product_in_stock ? 'default' : 'outline'}
                className="text-[10px] h-4 px-1.5 mt-0.5"
              >
                {entry.product_in_stock ? 'In Stock' : 'Out of Stock'}
              </Badge>
            </div>

            <span className="font-mono font-semibold shrink-0">
              {entry.product_currency} {entry.product_current_price}
            </span>

            <Button asChild variant="ghost" size="icon" className="shrink-0">
              <Link to={`/products/${entry.product_id}`}>
                <ExternalLink className="w-4 h-4" />
              </Link>
            </Button>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
