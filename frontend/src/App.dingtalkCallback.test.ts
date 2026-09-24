import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const appSource = readFileSync(new URL('./App.vue', import.meta.url), 'utf8')

test('global federated callback consumes DingTalk success and error fragments', () => {
  assert.match(appSource, /params\.get\(['"]dingtalk_result['"]\)/)
  assert.match(appSource, /params\.get\(['"]dingtalk_error['"]\)/)
  assert.match(appSource, /const encodedResult = oidcResult \|\| dingtalkResult/)
  assert.match(appSource, /handleGlobalOIDCCallback\(\)/)
})
