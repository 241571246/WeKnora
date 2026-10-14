import type { LocationQuery } from 'vue-router'
import { hasPendingFederatedCallback } from '../../router/federatedCallback'

/** Snapshot at setup: App.vue may consume/clear the callback hash on mount. */
export function getDingTalkAutoLoginEntry(entry: {
  path: string
  query: LocationQuery
  hash: string
}) {
  const query = { ...entry.query }
  const requested = entry.path === '/login'
    && query.autologindingtalk === 'true'
    && !hasPendingFederatedCallback(entry.hash)

  if (requested) delete query.autologindingtalk

  return {
    requested,
    // Invitation/register flows retain their existing priority and behavior.
    shouldLogin: requested && !query.token,
    query,
  }
}
