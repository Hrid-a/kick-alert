import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import api from '~/lib/api'

export function TrackForm() {
  const queryClient = useQueryClient()
  const [url, setUrl] = useState('')

  const trackMutation = useMutation({
    mutationFn: (productUrl: string) =>
      api.post('/v1/products', { product_url: productUrl }),
    onSuccess: () => {
      setUrl('')
      queryClient.invalidateQueries({ queryKey: ['watchlist'] })
      toast.success('Product queued', {
        description: 'It will appear in your watchlist shortly.',
      })
    },
    onError: (error: { response?: { data?: { error?: string } } }) => {
      toast.error(error?.response?.data?.error ?? 'Failed to track product.')
    },
  })

  return (
    <div className="space-y-2">
      <form
        onSubmit={(e) => {
          e.preventDefault()
          if (url.trim()) trackMutation.mutate(url.trim())
        }}
        className="flex gap-2"
      >
        <Input
          value={url}
          onChange={e => setUrl(e.target.value)}
          placeholder="Paste a product URL to track…"
          className="flex-1"
          disabled={trackMutation.isPending}
        />
        <Button type="submit" disabled={!url.trim() || trackMutation.isPending}>
          <Plus className="w-4 h-4 mr-1" /> Track
        </Button>
      </form>
    </div>
  )
}
