const FEDERATED_CALLBACK_KEYS = [
  'oidc_result',
  'oidc_error',
  'dingtalk_result',
  'dingtalk_error',
] as const

export function hasPendingFederatedCallback(hash: string | undefined) {
  if (!hash) return false

  const fragment = hash.startsWith('#') ? hash.slice(1) : hash
  const params = new URLSearchParams(fragment)
  return FEDERATED_CALLBACK_KEYS.some(key => params.has(key))
}
