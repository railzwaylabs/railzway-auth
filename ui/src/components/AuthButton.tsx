import type { ButtonHTMLAttributes } from 'react'

type AuthButtonVariant = 'primary' | 'secondary'

type AuthButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: AuthButtonVariant
}

const variantStyles: Record<AuthButtonVariant, string> = {
  primary:
    'bg-accent-primary text-text-inverse shadow-glow hover:bg-accent-primary/90 focus-visible:ring-accent-primary/40',
  secondary:
    'border border-border-subtle bg-bg-primary text-text-primary hover:border-accent-primary/40 hover:text-text-primary focus-visible:ring-accent-primary/30',
}

export default function AuthButton({
  variant = 'primary',
  className,
  type = 'button',
  ...props
}: AuthButtonProps) {
  const classes = [
    'inline-flex w-full items-center justify-center rounded-xl px-4 py-3 text-sm font-semibold transition duration-fast ease-standard',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-bg-surface',
    'disabled:cursor-not-allowed disabled:opacity-70',
    variantStyles[variant],
    className,
  ]
    .filter(Boolean)
    .join(' ')

  return <button type={type} className={classes} {...props} />
}
