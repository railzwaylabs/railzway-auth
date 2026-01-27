import { useMemo, useState } from 'react'
import { postJSON } from '../api'
import AuthBrand from '../components/AuthBrand'
import AuthButton from '../components/AuthButton'
import AuthCard from '../components/AuthCard'
import AuthInput from '../components/AuthInput'
import AuthLayout from '../components/AuthLayout'
import { buildQueryWithCurrent, getQueryParam } from '../utils/query'

type OTPVerifyResponse = {
  access_token: string
  refresh_token?: string
  token_type: string
  expires_in: number
  authorize_url?: string
}

export default function OTPVerify() {
  const returnTo = useMemo(() => getQueryParam('return_to') || '/', [])
  const authorizeState = useMemo(() => getQueryParam('state'), [])
  const sharedQuery = useMemo(() => buildQueryWithCurrent(), [])
  const scope = useMemo(() => getQueryParam('scope'), [])
  const initialClientId = useMemo(() => getQueryParam('client_id'), [])
  const presetIdentifier = useMemo(() => getQueryParam('phone'), [])

  const [identifier, setIdentifier] = useState(presetIdentifier)
  const [code, setCode] = useState('')
  const [clientId, setClientId] = useState(initialClientId)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const showClientIdInput = !clientId
  const requestQuery = useMemo(
    () => buildQueryWithCurrent({ phone: identifier }),
    [identifier],
  )

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)

    if (!clientId) {
      setError('Client ID is required to verify the OTP.')
      return
    }

    setSubmitting(true)
    try {
      const payload = await postJSON<OTPVerifyResponse>('/auth/otp/verify', {
        phone: identifier,
        code,
        client_id: clientId,
        scope: scope || undefined,
        state: authorizeState || undefined,
      })
      if (payload.authorize_url) {
        window.location.href = payload.authorize_url
        return
      }
      window.location.href = returnTo
    } catch (err) {
      setError(err instanceof Error ? err.message : 'OTP verification failed.')
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
              Verify your code
            </h1>
            <p className="text-sm text-text-muted">
              Enter the code sent to your email or phone.
            </p>
          </div>

          <form className="space-y-4" onSubmit={onSubmit}>
            <AuthInput
              label="Email or phone"
              type="text"
              autoComplete="username"
              required
              value={identifier}
              onChange={(event) => setIdentifier(event.target.value)}
            />

            <AuthInput
              label="One-time code"
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              required
              value={code}
              onChange={(event) => setCode(event.target.value)}
            />

            {showClientIdInput ? (
              <AuthInput
                label="Client ID"
                type="text"
                autoComplete="off"
                required
                helperText="Required for OTP verification"
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
              {submitting ? 'Verifying...' : 'Verify code'}
            </AuthButton>
          </form>

          <div className="flex items-center justify-between text-xs text-text-muted">
            <a
              className="text-text-secondary transition duration-fast ease-standard hover:text-text-primary"
              href={`/otp/request${requestQuery}`}
            >
              Send a new code
            </a>
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
