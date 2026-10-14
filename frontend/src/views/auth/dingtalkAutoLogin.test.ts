import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { getDingTalkAutoLoginEntry } from './dingtalkAutoLogin'

test('only exact opt-in on /login requests automatic DingTalk login', () => {
  assert.equal(getDingTalkAutoLoginEntry({ path: '/login', query: {}, hash: '' }).shouldLogin, false)
  for (const value of [null, '', 'false', '1', 'TRUE', ['true', 'true']]) {
    assert.equal(getDingTalkAutoLoginEntry({
      path: '/login', query: { autologindingtalk: value }, hash: '',
    }).shouldLogin, false)
  }
  assert.equal(getDingTalkAutoLoginEntry({
    path: '/login', query: { autologindingtalk: 'true' }, hash: '',
  }).shouldLogin, true)
  assert.equal(getDingTalkAutoLoginEntry({
    path: '/register', query: { autologindingtalk: 'true' }, hash: '',
  }).shouldLogin, false)
})

test('consumes the flag without mutating or dropping other query values', () => {
  const query = { autologindingtalk: 'true', h5: 'true', tag: ['a', 'b'] }
  const entry = getDingTalkAutoLoginEntry({ path: '/login', query, hash: '#section' })
  assert.deepEqual(entry.query, { h5: 'true', tag: ['a', 'b'] })
  assert.equal(query.autologindingtalk, 'true')
  assert.equal(getDingTalkAutoLoginEntry({
    path: '/login', query: entry.query, hash: '',
  }).shouldLogin, false, 'refresh after failure must not restart automatically')
})

test('pending success/error callbacks suppress auto login before App clears the hash', () => {
  for (const hash of ['#dingtalk_result=result', '#dingtalk_error=denied',
    '#oidc_result=result', '#oidc_error=denied']) {
    const entry = getDingTalkAutoLoginEntry({
      path: '/login', query: { autologindingtalk: 'true' }, hash,
    })
    assert.equal(entry.requested, false)
    assert.equal(entry.shouldLogin, false)
  }
})

test('invitation query is preserved but does not start a competing auth flow', () => {
  const entry = getDingTalkAutoLoginEntry({
    path: '/login', query: { autologindingtalk: 'true', token: 'invite-token' }, hash: '',
  })
  assert.equal(entry.requested, true)
  assert.equal(entry.shouldLogin, false)
  assert.deepEqual(entry.query, { token: 'invite-token' })
})

test('login integration consumes before async setup and awaits auth config before auto start', () => {
  const source = readFileSync(new URL('./Login.vue', import.meta.url), 'utf8')
  const mount = source.slice(source.indexOf('// Check if already logged in;'))
  assert.ok(mount.indexOf('query: autoDingTalkEntry.query') < mount.indexOf('await autoSetup()'))
  assert.ok(mount.indexOf('await loadFederatedAuthConfigs()') < mount.indexOf('await handleDingTalkLogin()'))
  assert.match(mount, /autoDingTalkEntry.shouldLogin && loginViewActive && route.path === '\/login'/)
  assert.match(source, /!dingtalkEnabled.value \|\| dingtalkLoading.value \|\| loading.value \|\| oidcLoading.value/)
})
