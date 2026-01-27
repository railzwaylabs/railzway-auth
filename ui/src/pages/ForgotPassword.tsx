import { useMemo, useState } from 'react'
import { postJSON } from '../api'
import AuthBrand from '../components/AuthBrand'
import AuthButton from '../components/AuthButton'
import AuthCard from '../components/AuthCard'
import AuthInput from '../components/AuthInput'
import AuthLayout from '../components/AuthLayout'
import { buildQueryWithCurrent, getQueryParam } from '../utils/query'

type ForgotResponse = {
  message?: string
}

export default function ForgotPassword() {
  const sharedQuery = useMemo(() => buildQueryWithCurrent(), [])
  const presetEmail = useMemo(() => getQueryParam('email'), [])

  const [email, setEmail] = useState(presetEmail)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setSubmitting(true)

    try {
      const payload = await postJSON<ForgotResponse>('/auth/password/forgot', {
        email,
      })
      setSuccess(
        payload.message ||
          'If the account exists, password reset instructions have been sent.',
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Request failed.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout>
      <AuthCard>
        <div className="space-y-6">
          <AuthBrand />

          <div className="space-y-2">
            <h1 className="text-3xl font-semibold tracking-tight text-text-primary">
              Reset your password
            </h1>
            <p className="text-sm text-text-muted">
              We will email reset instructions if the account exists.
            </p>
          </div>

          <form className="space-y-4" onSubmit={onSubmit}>
            <AuthInput
              label="Email"
              type="email"
              autoComplete="email"
              required
              value={email}
              onChange={(event) => setEmail(event.target.value)}
            />

            {success ? (
              <div className="rounded-xl border border-border-subtle bg-bg-surface px-4 py-3 text-sm text-text-secondary">
                {success}
              </div>
            ) : null}

            {error ? (
              <div
                className="rounded-xl border border-status-error/40 bg-status-error/10 px-4 py-3 text-sm text-status-error"
                role="alert"
              >
                {error}
              </div>
            ) : null}

            <AuthButton type="submit" disabled={submitting}>
              {submitting ? 'Sending...' : 'Send reset email'}
            </AuthButton>
          </form>

          <div className="text-xs text-text-muted">
            <a
              className="text-text-secondary transition duration-fast ease-standard hover:text-text-primary"
              href={`/login${sharedQuery}`}
            >
              Back to sign in
            </a>
          </div>
        </div>
      </AuthCard>
    </AuthLayout>
  )
}
