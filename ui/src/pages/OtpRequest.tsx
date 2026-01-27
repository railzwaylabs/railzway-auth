import { useMemo, useState } from 'react'
import { postJSON } from '../api'
import AuthBrand from '../components/AuthBrand'
import AuthButton from '../components/AuthButton'
import AuthCard from '../components/AuthCard'
import AuthInput from '../components/AuthInput'
import AuthLayout from '../components/AuthLayout'
import { buildQueryWithCurrent, getQueryParam } from '../utils/query'

type OTPRequestResponse = {
  message?: string
}

export default function OTPRequest() {
  const authorizeState = useMemo(() => getQueryParam('state'), [])
  const sharedQuery = useMemo(() => buildQueryWithCurrent(), [])
  const presetIdentifier = useMemo(() => getQueryParam('phone'), [])

  const [identifier, setIdentifier] = useState(presetIdentifier)
  const [channel, setChannel] = useState('sms')
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setSubmitting(true)

    try {
      const payload = await postJSON<OTPRequestResponse>('/auth/otp/request', {
        phone: identifier,
        channel,
        state: authorizeState || undefined,
      })
      setSuccess(payload.message || 'OTP request accepted.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'OTP request failed.')
    } finally {
      setSubmitting(false)
    }
  }

  const verifyQuery = useMemo(
    () => buildQueryWithCurrent({ phone: identifier }),
    [identifier],
  )

  return (
    <AuthLayout>
      <AuthCard>
        <div className="space-y-6">
          <AuthBrand />

          <div className="space-y-2">
            <h1 className="text-3xl font-semibold tracking-tight text-text-primary">
              Request a one-time code
            </h1>
            <p className="text-sm text-text-muted">
              We will send a short-lived code to continue signing in.
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

            <label className="block space-y-2 text-sm">
              <span className="font-medium text-text-secondary">
                Delivery channel
              </span>
              <select
                className="w-full rounded-2xl border border-border-subtle/70 bg-bg-surface px-4 py-3.5 text-base text-text-primary outline-none transition duration-fast ease-standard focus:border-text-primary/70 focus:ring-2 focus:ring-text-primary/15"
                value={channel}
                onChange={(event) => setChannel(event.target.value)}
              >
                <option value="sms">SMS</option>
                <option value="email">Email</option>
              </select>
            </label>

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
              {submitting ? 'Requesting...' : 'Send code'}
            </AuthButton>
          </form>

          <div className="flex items-center justify-between text-xs text-text-muted">
            <a
              className="text-text-secondary transition duration-fast ease-standard hover:text-text-primary"
              href={`/otp/verify${verifyQuery}`}
            >
              Already have a code?
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
