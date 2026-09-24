import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { normalizeExternalSystemUrl } from './externalSystemUrl'

describe('normalizeExternalSystemUrl', () => {
  it('accepts absolute http and https URLs', () => {
    assert.equal(normalizeExternalSystemUrl('https://pms.example.com/app'), 'https://pms.example.com/app')
    assert.equal(normalizeExternalSystemUrl(' http://localhost:5173 '), 'http://localhost:5173/')
  })

  it('rejects URLs that could execute or inherit local routing', () => {
    assert.equal(normalizeExternalSystemUrl(undefined), null)
    assert.equal(normalizeExternalSystemUrl('/login'), null)
    assert.equal(normalizeExternalSystemUrl('javascript:alert(1)'), null)
    assert.equal(normalizeExternalSystemUrl('data:text/html,hello'), null)
  })
})
