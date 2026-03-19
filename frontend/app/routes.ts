import { type RouteConfig, index, layout, route } from '@react-router/dev/routes'

export default [
  index('routes/home.tsx'),

  layout('routes/_auth.tsx', [
    route('login', 'routes/_auth.login.tsx'),
    route('register', 'routes/_auth.register.tsx'),
    route('activate', 'routes/_auth.activate.tsx'),
  ]),

  layout('routes/_protected.tsx', [
    route('dashboard', 'routes/_protected.dashboard.tsx'),
    route('products', 'routes/_protected.products.tsx'),
    route('products/:id', 'routes/_protected.products.$id.tsx'),
    route('notifications', 'routes/_protected.notifications.tsx'),
    route('settings', 'routes/_protected.settings.tsx'),
  ]),
] satisfies RouteConfig
