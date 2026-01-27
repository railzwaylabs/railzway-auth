import { useMemo, useState } from 'react'
import { postJSON } from '../api'
import AuthBrand from '../components/AuthBrand'
import AuthButton from '../components/AuthButton'
import AuthCard from '../components/AuthCard'
import AuthInput from '../components/AuthInput'
import AuthLayout from '../components/AuthLayout'
import {
  passwordInputPattern,
  passwordRequirements,
  validatePassword,
} from '../utils/password'
import { buildQueryWithCurrent, getQueryParam } from '../utils/query'

type RegisterResponse = {
  access_token: string
  refresh_token?: string
  token_type: string
  expires_in: number
  authorize_url?: string
}

export default function Register() {
  const returnTo = useMemo(() => getQueryParam('return_to') || '/', [])
  const authorizeState = useMemo(() => getQueryParam('state'), [])
  const sharedQuery = useMemo(() => buildQueryWithCurrent(), [])
  const initialClientId = useMemo(() => getQueryParam('client_id'), [])

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [clientId, setClientId] = useState(initialClientId)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const showClientIdInput = !authorizeState && !clientId

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)

    if (!authorizeState && !clientId) {
      setError('Missing state or client ID. Please start from the OAuth flow.')
      return
    }

    const passwordError = validatePassword(password)
    if (passwordError) {
      setError(passwordError)
      return
    }

    setSubmitting(true)
    try {
      const payload = await postJSON<RegisterResponse>('/auth/password/register', {
        email,
        password,
        name,
        client_id: clientId || undefined,
        state: authorizeState || undefined,
      })
      if (payload.authorize_url) {
        window.location.href = payload.authorize_url
        return
      }
      window.location.href = returnTo
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed.')
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
              Create account
            </h1>
            <p className="text-sm text-text-muted">
              Use your work email to join Railzway Cloud.
            </p>
          </div>

          <form className="space-y-4" onSubmit={onSubmit}>
            <AuthInput
              label="Full name"
              type="text"
              autoComplete="name"
              helperText="Optional"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />

            <AuthInput
              label="Email"
              type="email"
              autoComplete="email"
              required
              value={email}
              onChange={(event) => setEmail(event.target.value)}
            />

            <AuthInput
              label="Password"
              type="password"
              autoComplete="new-password"
              helperText={passwordRequirements()}
              minLength={8}
              pattern={passwordInputPattern}
              title={passwordRequirements()}
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />

            {showClientIdInput ? (
              <AuthInput
                label="Client ID"
                type="text"
                autoComplete="off"
                required
                helperText="Provided by your organization"
                value={clientId}
                onChange={(event) => setClientId(event.target.value)}
              />
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
              {submitting ? 'Creating account...' : 'Create account'}
            </AuthButton>
          </form>

          <div className="text-xs text-text-muted">
            Already have an account?{' '}
            <a
              className="text-text-muted underline-offset-4 transition duration-fast ease-standard hover:text-text-secondary hover:underline"
              href={`/login${sharedQuery}`}
            >
              Sign in
            </a>
          </div>
        </div>
      </AuthCard>
    </AuthLayout>
  )
}
