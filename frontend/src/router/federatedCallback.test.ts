import assert from 'node:assert/strict'
import test from 'node:test'

import { hasPendingFederatedCallback } from './federatedCallback'

test('recognizes OIDC callback fragments', () => {
  assert.equal(hasPendingFederatedCallback('#oidc_result=encoded'), true)
  assert.equal(hasPendingFederatedCallback('#oidc_error=access_denied'), true)
})

test('recognizes DingTalk callback fragments', () => {
  assert.equal(hasPendingFederatedCallback('#dingtalk_result=encoded'), true)
  assert.equal(hasPendingFederatedCallback('#dingtalk_error=login_failed'), true)
})

test('ignores unrelated or embedded callback text', () => {
  assert.equal(hasPendingFederatedCallback(''), false)
  assert.equal(hasPendingFederatedCallback('#section=login'), false)
  assert.equal(hasPendingFederatedCallback('#message=dingtalk_result%3Dencoded'), false)
})
