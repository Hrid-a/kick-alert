import { useMutation } from '@tanstack/react-query'
import { Link, useNavigate, useSearchParams } from 'react-router'
import { CheckCircle, XCircle, Loader2 } from 'lucide-react'
import api from '~/lib/api'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'

export default function ActivatePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')

  const mutation = useMutation({
    mutationFn: (t: string) => api.put('/v1/activation', { token: t }),
    onSuccess: () => {
      setTimeout(() => navigate('/login', { replace: true }), 2000)
    },
  })

  const handleActivation = ()=>{
    if (token) mutation.mutate(token)
  }


  return (
    <Card className="w-full max-w-md bg-card border-border text-center">
      <CardHeader className="space-y-2">
        <CardTitle className="text-2xl font-display uppercase tracking-tight">KickAlert</CardTitle>
        <CardDescription>Account Activation</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!token && (
          <p className="text-muted-foreground text-sm">
            Check your email for an activation link.
          </p>
        )}

        {token && mutation.isIdle && (
          <div className="flex flex-col items-center gap-3 py-4">
            <Button onClick={handleActivation}>Activate your email</Button>
          </div>
        )}

        {token && mutation.isPending && (
          <div className="flex flex-col items-center gap-3 py-4">
            <Loader2 className="w-8 h-8 animate-spin text-accent" />
            <p className="text-muted-foreground text-sm">Activating your account…</p>
          </div>
        )}

        {mutation.isSuccess && (
          <div className="flex flex-col items-center gap-3 py-4">
            <CheckCircle className="w-8 h-8 text-accent" />
            <p className="text-sm">Account activated! Redirecting to login…</p>
          </div>
        )}

        {mutation.isError && (
          <div className="flex flex-col items-center gap-3 py-4">
            <XCircle className="w-8 h-8 text-destructive" />
            <p className="text-sm text-destructive">
              {(mutation.error as { response?: { data?: { error?: string } } })?.response?.data
                ?.error ?? 'Activation failed. The link may have expired.'}
            </p>
            <Button asChild variant="outline" size="sm">
              <Link to="/register">Back to Sign Up</Link>
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
