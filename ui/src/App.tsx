import AuthBrand from './components/AuthBrand'
import AuthButton from './components/AuthButton'
import AuthCard from './components/AuthCard'
import AuthLayout from './components/AuthLayout'
import ErrorPage from './pages/Error'
import ForgotPassword from './pages/ForgotPassword'
import Login from './pages/Login'
import OTPRequest from './pages/OtpRequest'
import OTPVerify from './pages/OtpVerify'
import Register from './pages/Register'

function App() {
  const path = window.location.pathname

  if (path === '/' || path === '/login') {
    return <Login />
  }

  if (path === '/register') {
    return <Register />
  }

  if (path === '/forgot-password') {
    return <ForgotPassword />
  }

  if (path === '/otp/request') {
    return <OTPRequest />
  }

  if (path === '/otp/verify') {
    return <OTPVerify />
  }

  if (path === '/error') {
    return <ErrorPage />
  }

  return (
    <AuthLayout>
      <AuthCard>
        <div className="space-y-6">
          <AuthBrand />
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold text-text-primary">
              Page not found
            </h1>
            <p className="text-sm text-text-muted">
              The page you are looking for does not exist.
            </p>
          </div>
          <AuthButton onClick={() => (window.location.href = '/login')}>
            Go to login
          </AuthButton>
        </div>
      </AuthCard>
    </AuthLayout>
  )
}

export default App
