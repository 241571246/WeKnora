export function normalizeExternalSystemUrl(value: string | undefined): string | null {
  const candidate = value?.trim()
  if (!candidate) return null
  try {
    const url = new URL(candidate)
    if (url.protocol !== 'http:' && url.protocol !== 'https:') return null
    return url.toString()
  } catch {
    return null
  }
}

export function configuredPMSUrl(): string | null {
  return normalizeExternalSystemUrl(
    window.__RUNTIME_CONFIG__?.PMS_URL ?? import.meta.env.VITE_PMS_URL,
  )
}
