import { useParams, Link } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import { ArrowLeft, ExternalLink, Package, TrendingDown } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '~/components/ui/card'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Skeleton } from '~/components/ui/skeleton'
import api from '~/lib/api'

type Product = {
  slug: string
  name: string
  sku: string
  cloudProductId: string
  category: string
  url: string
  image_url: string
  current_price: string
  currency: string
  in_stock: boolean
  last_scraped_at: string
}

type PricePoint = {
  price: string
  in_stock: boolean
  scraped_at: string
}

function ProductSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-48 w-full rounded-xl" />
      <Skeleton className="h-72 w-full rounded-xl" />
    </div>
  )
}

function PriceChart({ data, currency }: { data: PricePoint[]; currency: string }) {
  const chartData = data.map(p => ({
    date: new Date(p.scraped_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    price: parseFloat(p.price),
    in_stock: p.in_stock,
  }))

  const prices = chartData.map(d => d.price)
  const min = Math.min(...prices)
  const max = Math.max(...prices)
  const padding = (max - min) * 0.1 || 1

  return (
    <ResponsiveContainer width="100%" height={260}>
      <AreaChart data={chartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
        <defs>
          <linearGradient id="priceGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#4ade80" stopOpacity={0.25} />
            <stop offset="95%" stopColor="#4ade80" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="#262626" vertical={false} />
        <XAxis
          dataKey="date"
          tick={{ fill: '#a0a0a0', fontSize: 11 }}
          axisLine={false}
          tickLine={false}
          interval="preserveStartEnd"
        />
        <YAxis
          domain={[min - padding, max + padding]}
          tick={{ fill: '#a0a0a0', fontSize: 11 }}
          axisLine={false}
          tickLine={false}
          tickFormatter={v => `${currency}${v.toFixed(0)}`}
          width={60}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: '#1a1a1a',
            border: '1px solid #262626',
            borderRadius: '8px',
            color: '#fff',
            fontSize: 12,
          }}
          formatter={(value) => [`${currency}${Number(value).toFixed(2)}`, 'Price']}
          labelStyle={{ color: '#a0a0a0' }}
        />
        <Area
          type="monotone"
          dataKey="price"
          stroke="#4ade80"
          strokeWidth={2}
          fill="url(#priceGradient)"
          dot={false}
          activeDot={{ r: 4, fill: '#4ade80' }}
        />
      </AreaChart>
    </ResponsiveContainer>
  )
}

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>()

  const { data: product, isLoading: productLoading, isError } = useQuery<Product>({
    queryKey: ['product', id],
    queryFn: () => api.get(`/v1/products/${id}`).then(r => r.data.product),
    enabled: !!id,
  })

  const { data: history, isLoading: historyLoading } = useQuery<PricePoint[]>({
    queryKey: ['price-history', id],
    queryFn: () => api.get(`/v1/products/${id}/price-history`).then(r => r.data.price_history),
    enabled: !!id,
  })

  const lowestPrice = history?.length
    ? Math.min(...history.map(p => parseFloat(p.price)))
    : null

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-3xl mx-auto px-6 py-10 space-y-6">
        {/* Back */}
        <Button asChild variant="ghost" size="sm" className="-ml-2">
          <Link to="/dashboard">
            <ArrowLeft className="w-4 h-4 mr-1" /> Back
          </Link>
        </Button>

        {productLoading && <ProductSkeleton />}

        {isError && (
          <p className="text-sm text-destructive">Product not found.</p>
        )}

        {product && (
          <>
            {/* Product card */}
            <Card>
              <CardContent className="flex gap-6 py-6">
                {/* Image */}
                {product.image_url ? (
                  <img
                    src={product.image_url}
                    alt={product.name}
                    className="w-28 h-28 object-cover rounded-lg shrink-0"
                  />
                ) : (
                  <div className="w-28 h-28 bg-muted rounded-lg flex items-center justify-center shrink-0">
                    <Package className="w-8 h-8 text-muted-foreground" />
                  </div>
                )}

                {/* Info */}
                <div className="flex-1 min-w-0 space-y-3">
                  <div>
                    <h1 className="text-xl font-bold leading-tight">{product.name}</h1>
                    <p className="text-xs text-muted-foreground mt-0.5">SKU: {product.sku}</p>
                  </div>

                  <div className="flex items-center gap-3 flex-wrap">
                    <span className="text-2xl font-mono font-bold">
                      {product.currency} {product.current_price}
                    </span>
                    <Badge variant={product.in_stock ? 'default' : 'outline'}>
                      {product.in_stock ? 'In Stock' : 'Out of Stock'}
                    </Badge>
                  </div>

                  <div className="flex items-center gap-3 flex-wrap text-xs text-muted-foreground">
                    <span className="capitalize">{product.category.toLowerCase()}</span>
                    {lowestPrice !== null && (
                      <span className="flex items-center gap-1 text-accent">
                        <TrendingDown className="w-3 h-3" />
                        Lowest: {product.currency} {lowestPrice.toFixed(2)}
                      </span>
                    )}
                  </div>

                  {product.url && (
                    <Button asChild variant="outline" size="sm" className="w-fit">
                      <a href={product.url} target="_blank" rel="noopener noreferrer">
                        <ExternalLink className="w-3 h-3 mr-1.5" /> View Product
                      </a>
                    </Button>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Price history */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-base">Price History</CardTitle>
              </CardHeader>
              <CardContent>
                {historyLoading && <Skeleton className="h-64 w-full rounded-lg" />}

                {!historyLoading && (!history || history.length === 0) && (
                  <div className="h-40 flex items-center justify-center">
                    <p className="text-sm text-muted-foreground">
                      No price history recorded yet.
                    </p>
                  </div>
                )}

                {!historyLoading && history && history.length > 0 && (
                  <PriceChart data={history} currency={product.currency} />
                )}
              </CardContent>
            </Card>
          </>
        )}
      </div>
    </div>
  )
}
