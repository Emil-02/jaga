import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { installHook, uninstallHook } from '../src/hook.mjs'
import { tempRepo } from './helpers.mjs'

test('hook dipasang sekali dan bisa dilepas tanpa merusak isi lama', async () => {
  const root = await tempRepo(['.git/hooks'], {
    '.git/hooks/pre-commit': '#!/bin/sh\necho "custom"\n',
  })
  await installHook(root)
  await installHook(root)
  let hook = await readFile(join(root, '.git/hooks/pre-commit'), 'utf8')
  assert.equal(hook.split('# jaga:begin').length - 1, 1)
  assert.ok(hook.includes('echo "custom"'))

  await uninstallHook(root)
  hook = await readFile(join(root, '.git/hooks/pre-commit'), 'utf8')
  assert.ok(hook.includes('echo "custom"'))
  assert.ok(!hook.includes('jaga:begin'))
})

test('hook membuat file baru bila belum ada', async () => {
  const root = await tempRepo(['.git/hooks'])
  const result = await installHook(root)
  assert.equal(result.ok, true)
  const hook = await readFile(join(root, '.git/hooks/pre-commit'), 'utf8')
  assert.ok(hook.startsWith('#!/bin/sh'))
  assert.ok(hook.includes('jaga-skill check'))
})
