import { Toaster as Sonner } from 'sonner'
import type { ToasterProps } from 'sonner'

const Toaster = ({ ...props }: ToasterProps) => {
  return (
    <Sonner
      theme="dark"
      className="toaster group"
      style={
        {
          '--normal-bg': 'var(--popover)',
          '--normal-text': 'var(--popover-foreground)',
          '--normal-border': 'var(--border)',
          '--success-bg': '#052e16',
          '--success-text': '#4ade80',
          '--success-border': '#166534',
          '--error-bg': '#450a0a',
          '--error-text': '#ef4444',
          '--error-border': '#991b1b',
        } as React.CSSProperties
      }
      {...props}
    />
  )
}

export { Toaster }
