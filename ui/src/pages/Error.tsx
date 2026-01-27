import AuthBrand from '../components/AuthBrand'
import AuthButton from '../components/AuthButton'
import AuthCard from '../components/AuthCard'
import AuthLayout from '../components/AuthLayout'

export default function ErrorPage() {
  const params = new URLSearchParams(window.location.search)
  const message =
    params.get('error_description') ||
    params.get('error') ||
    'Something went wrong.'

  return (
    <AuthLayout>
      <AuthCard>
        <div className="space-y-6">
          <AuthBrand />
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold text-text-primary">Error</h1>
            <p className="text-sm text-text-muted">{message}</p>
          </div>
          <AuthButton onClick={() => (window.location.href = '/login')}>
            Back to login
          </AuthButton>
        </div>
      </AuthCard>
    </AuthLayout>
  )
}
