import test from 'node:test'
import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { tempRepo } from './helpers.mjs'

const scanner = fileURLToPath(new URL('../src/js-scanner.mjs', import.meta.url))

test('fallback JS mendeteksi secret dan menyembunyikan nilainya', async () => {
  const root = await tempRepo([], {
    'config.js': "export const key = 'AKIA1234567890ABCDEF'\n",
  })
  const res = spawnSync(process.execPath, [scanner, '--format', 'text', root], { encoding: 'utf8' })
  assert.equal(res.status, 1)
  assert.ok(res.stdout.includes('AWS access key'))
  assert.ok(!res.stdout.includes('AKIA1234567890ABCDEF'))
  assert.ok(res.stdout.includes('AKIA…CDEF'))
})

test('fallback JS lolos pada repo bersih', async () => {
  const root = await tempRepo([], { 'README.md': '# Halo\n' })
  const res = spawnSync(process.execPath, [scanner, '--format', 'sarif', root], { encoding: 'utf8' })
  assert.equal(res.status, 0)
  const doc = JSON.parse(res.stdout)
  assert.equal(doc.version, '2.1.0')
  assert.equal(doc.runs[0].results.length, 0)
})
