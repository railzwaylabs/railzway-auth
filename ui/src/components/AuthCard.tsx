import type { ReactNode } from 'react'

type AuthCardProps = {
  children: ReactNode
}

export default function AuthCard({ children }: AuthCardProps) {
  return (
    <div className="rounded-3xl border border-border-subtle bg-bg-surface/90 p-8 shadow-md ring-1 ring-border-subtle/60 backdrop-blur-sm motion-safe:animate-slide-up">
      {children}
    </div>
  )
}
