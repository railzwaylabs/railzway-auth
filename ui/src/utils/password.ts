const passwordPattern = /^(?=.*[A-Z])(?=.*[^A-Za-z0-9]).{8,}$/

export function validatePassword(value: string): string | null {
  if (passwordPattern.test(value)) {
    return null
  }
  return 'Password must be at least 8 characters and include 1 uppercase letter and 1 special character.'
}

export function passwordRequirements(): string {
  return 'At least 8 characters, 1 uppercase letter, and 1 special character.'
}

export const passwordInputPattern = passwordPattern.source
