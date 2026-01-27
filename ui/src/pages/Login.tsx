import { useEffect, useMemo, useState } from 'react'
import { getJSON, postJSON } from '../api'
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

type LoginResponse = {
  access_token: string
  refresh_token?: string
  token_type: string
  expires_in: number
  authorize_url?: string
}

type OAuthProvider = {
  Name: string
  DisplayName?: string
  IconURL?: string
}

type OAuthStartResponse = {
  authorization_url: string
}

function formatProviderLabel(provider: OAuthProvider) {
  if (provider.DisplayName && provider.DisplayName.trim()) {
    return provider.DisplayName
  }
  const raw = provider.Name || ''
  return raw ? raw.charAt(0).toUpperCase() + raw.slice(1) : 'Provider'
}

export default function Login() {
  const returnTo = useMemo(() => getQueryParam('return_to') || '/', [])
  const authorizeState = useMemo(() => getQueryParam('state'), [])
  const scope = useMemo(
    () => getQueryParam('scope') || 'openid email profile',
    [],
  )
  const sharedQuery = useMemo(() => buildQueryWithCurrent(), [])

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [providers, setProviders] = useState<OAuthProvider[]>([])

  useEffect(() => {
    let active = true
    async function loadProviders() {
      try {
        const payload = await getJSON<OAuthProvider[]>('/auth/oauth/providers')
        if (active) {
          setProviders(payload)
        }
      } catch {
        if (active) {
          setProviders([])
        }
      }
    }

    loadProviders()

    return () => {
      active = false
    }
  }, [])

  async function startOAuth(provider: OAuthProvider) {
    setError(null)
    try {
      const callbackURL = new URL('/auth/oauth/callback', window.location.origin)
      callbackURL.searchParams.set('provider', provider.Name)
      if (returnTo) {
        callbackURL.searchParams.set('redirect_uri', returnTo)
      }

      const startURL = new URL('/auth/oauth/start', window.location.origin)
      startURL.searchParams.set('provider', provider.Name)
      startURL.searchParams.set('redirect_uri', callbackURL.toString())
      if (scope) {
        startURL.searchParams.set('scope', scope)
      }

      const payload = await getJSON<OAuthStartResponse>(startURL.toString())
      window.location.href = payload.authorization_url
    } catch (err) {
      setError(err instanceof Error ? err.message : 'OAuth start failed.')
    }
  }

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)

    if (!authorizeState) {
      setError('Missing state. Please retry from the OAuth authorize flow.')
      return
    }

    const passwordError = validatePassword(password)
    if (passwordError) {
      setError(passwordError)
      return
    }

    setSubmitting(true)
    try {
      // UI never stores tokens; the backend sets HttpOnly session cookies.
      const payload = await postJSON<LoginResponse>('/auth/password/login', {
        email,
        password,
        scope,
        state: authorizeState,
      })
      if (payload.authorize_url) {
        window.location.href = payload.authorize_url
        return
      }
      window.location.href = returnTo
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed.')
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
              Sign in
            </h1>
            <p className="text-sm text-text-muted">
              Use your Railzway Cloud credentials to continue.
            </p>
          </div>

          <form className="space-y-5" onSubmit={onSubmit}>
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
              labelAction={
                <a
                  className="text-text-muted transition duration-fast ease-standard hover:text-text-secondary"
                  href={`/forgot-password${sharedQuery}`}
                >
                  Forgot your password?
                </a>
              }
              type="password"
              autoComplete="current-password"
              minLength={8}
              pattern={passwordInputPattern}
              title={passwordRequirements()}
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />

            {error ? (
              <div
                className="rounded-xl border border-status-error/40 bg-status-error/10 px-4 py-3 text-sm text-status-error"
                role="alert"
              >
                {error}
              </div>
            ) : null}

            <AuthButton type="submit" disabled={submitting}>
              {submitting ? 'Signing in...' : 'Sign in'}
            </AuthButton>
          </form>

          {providers.length > 0 ? (
            <div className="space-y-4">
              <div className="flex items-center gap-3 text-xs text-text-muted">
                <span className="h-px flex-1 bg-border-subtle/60" />
                <span>Or continue with</span>
                <span className="h-px flex-1 bg-border-subtle/60" />
              </div>
              <div className="space-y-3">
                {providers.map((provider) => (
                  <AuthButton
                    key={provider.Name}
                    variant="secondary"
                    className="justify-center gap-2"
                    onClick={() => startOAuth(provider)}
                    type="button"
                  >
                    {provider.IconURL ? (
                      <img
                        src={provider.IconURL}
                        alt={formatProviderLabel(provider)}
                        className="h-5 w-5"
                      />
                    ) : null}
                    Login with {formatProviderLabel(provider)}
                  </AuthButton>
                ))}
              </div>
            </div>
          ) : null}

          <div className="flex flex-col gap-3 text-xs text-text-muted">
            <a
              className="text-text-muted transition duration-fast ease-standard hover:text-text-secondary"
              href={`/otp/request${sharedQuery}`}
            >
              Use a one-time code instead
            </a>
            <span>
              Don&apos;t have an account?{' '}
              <a
                className="text-text-muted underline-offset-4 transition duration-fast ease-standard hover:text-text-secondary hover:underline"
                href={`/register${sharedQuery}`}
              >
                Sign up
              </a>
            </span>
          </div>
        </div>
      </AuthCard>
    </AuthLayout>
  )
}
