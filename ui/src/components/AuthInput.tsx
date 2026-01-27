import { useId } from 'react'
import type { InputHTMLAttributes, ReactNode } from 'react'

type AuthInputProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string
  helperText?: string
  error?: string
  labelAction?: ReactNode
}

export default function AuthInput({
  label,
  helperText,
  error,
  labelAction,
  className,
  id,
  ...props
}: AuthInputProps) {
  const inputId = id ?? useId()
  const inputClasses = [
    'w-full rounded-xl border border-border-subtle bg-bg-primary px-4 py-3 text-base text-text-primary outline-none transition duration-fast ease-standard',
    'placeholder:text-text-muted/70',
    error
      ? 'border-status-error/60 focus:border-status-error focus:ring-2 focus:ring-status-error/30'
      : 'focus:border-accent-primary focus:ring-2 focus:ring-accent-primary/40',
    className,
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <label className="block space-y-2 text-sm" htmlFor={inputId}>
      <span className="flex items-center justify-between gap-3">
        <span className="font-medium text-text-secondary">{label}</span>
        {labelAction ? (
          <span className="text-xs text-text-muted">{labelAction}</span>
        ) : null}
      </span>
      <input
        id={inputId}
        className={inputClasses}
        aria-invalid={Boolean(error)}
        {...props}
      />
      {helperText ? (
        <span className="block text-xs text-text-muted">{helperText}</span>
      ) : null}
      {error ? (
        <span className="block text-xs text-status-error">{error}</span>
      ) : null}
    </label>
  )
}
