const preservedKeys = ['state', 'return_to', 'client_id', 'scope'] as const

type QueryValue = string | null | undefined

type QueryMap = Record<string, QueryValue>

export function getQueryParam(name: string): string {
  const params = new URLSearchParams(window.location.search)
  return params.get(name) ?? ''
}

export function buildQueryWithCurrent(overrides: QueryMap = {}): string {
  const current = new URLSearchParams(window.location.search)
  const next = new URLSearchParams()

  preservedKeys.forEach((key) => {
    const value = current.get(key)
    if (value) {
      next.set(key, value)
    }
  })

  Object.entries(overrides).forEach(([key, value]) => {
    if (value) {
      next.set(key, value)
    }
  })

  const query = next.toString()
  return query ? `?${query}` : ''
}
