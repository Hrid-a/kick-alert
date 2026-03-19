import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { Link, useNavigate } from 'react-router'
import { useState } from 'react'
import { Eye, EyeOff } from 'lucide-react'
import { loginSchema, type LoginInput } from '~/lib/schema'
import { useAuth } from '~/context/auth'
import api from '~/lib/api'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import { Spinner } from '~/components/ui/spinner'

export default function LoginPage() {
  const navigate = useNavigate()
  const { setAuth } = useAuth()
  const [showPassword, setShowPassword] = useState(false)

  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  })

  const mutation = useMutation({
    mutationFn: (data: LoginInput) =>
      api.post<{ user: { id: string; name: string; email: string }; token: string }>('/v1/login', data).then((r) => r.data),
    onSuccess: (data) => {
      setAuth(data.token, data.user)
      navigate('/dashboard', { replace: true })
    },
    onError: (err: unknown) => {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Login failed. Please try again.'
      form.setError('root', { message })
    },
  })

  return (
    <Card className="w-full max-w-md bg-card border-border">
      <CardHeader className="space-y-2 text-center">
        <CardTitle className="text-2xl font-display uppercase tracking-tight">KickAlert</CardTitle>
        <CardDescription>Track Nike prices, get alerts</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={form.handleSubmit((d) => mutation.mutate(d))}
          className="space-y-4"
          noValidate
        >
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground">Email</label>
            <Input
              type="email"
              placeholder="you@example.com"
              autoComplete="email"
              disabled={mutation.isPending}
              aria-invalid={!!form.formState.errors.email}
              className="bg-input border-border"
              {...form.register('email')}
            />
            {form.formState.errors.email && (
              <p className="text-sm text-destructive">{form.formState.errors.email.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground">Password</label>
            <div className="relative">
              <Input
                type={showPassword ? 'text' : 'password'}
                placeholder="••••••••"
                autoComplete="current-password"
                disabled={mutation.isPending}
                aria-invalid={!!form.formState.errors.password}
                className="bg-input border-border pr-10"
                {...form.register('password')}
              />
              <button
                type="button"
                onClick={() => setShowPassword((p) => !p)}
                className="absolute inset-y-0 end-3 flex items-center text-muted-foreground hover:text-foreground"
                tabIndex={-1}
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
            {form.formState.errors.password && (
              <p className="text-sm text-destructive">{form.formState.errors.password.message}</p>
            )}
          </div>

          {form.formState.errors.root && (
            <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md text-sm text-destructive">
              {form.formState.errors.root.message}
            </div>
          )}

          <Button
            type="submit"
            disabled={mutation.isPending}
            className="w-full bg-accent hover:bg-accent/90 text-accent-foreground font-medium"
          >
            {mutation.isPending ? (
              <>
                <Spinner className="w-4 h-4 mr-2" />
                Signing in...
              </>
            ) : (
              'Sign In'
            )}
          </Button>
        </form>

        <div className="mt-6 text-center text-sm">
          <span className="text-muted-foreground">Don&apos;t have an account? </span>
          <Link to="/register" className="text-accent hover:text-accent/80 font-medium transition">
            Sign up
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
