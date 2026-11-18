import test from 'node:test'
import assert from 'node:assert/strict'
import { detectAgents } from '../src/detect.mjs'
import { tempRepo } from './helpers.mjs'

test('deteksi beberapa agent sekaligus', async () => {
  const root = await tempRepo(['.opencode', '.claude', '.codex'])
  const ids = (await detectAgents(root)).map((t) => t.id)
  assert.deepEqual(ids.sort(), ['claude', 'codex', 'opencode'])
})

test('fallback generic bila tidak ada agent', async () => {
  const root = await tempRepo()
  const ids = (await detectAgents(root)).map((t) => t.id)
  assert.deepEqual(ids, ['generic'])
})

test('gemini terdeteksi dari GEMINI.md', async () => {
  const root = await tempRepo([], { 'GEMINI.md': '# hello' })
  const ids = (await detectAgents(root)).map((t) => t.id)
  assert.ok(ids.includes('gemini'))
})
