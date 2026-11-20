import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { install } from '../src/install.mjs'
import { uninstall } from '../src/uninstall.mjs'
import { tempRepo } from './helpers.mjs'

const pkgRoot = fileURLToPath(new URL('..', import.meta.url))

test('uninstall hanya menghapus file yang dikelola', async () => {
  const root = await tempRepo(['.opencode'], {
    'AGENTS.md': '# Repo saya\n\nKonten penting user tetap ada.\n',
    'src/app.js': 'console.log(1)\n',
  })
  await install(root, { pkgRoot })
  const result = await uninstall(root)
  assert.ok(result.removed.includes('.opencode/skills/jaga/SKILL.md'))
  const agents = await readFile(join(root, 'AGENTS.md'), 'utf8')
  assert.ok(agents.includes('# Repo saya'))
  assert.ok(agents.includes('Konten penting user tetap ada.'))
  assert.ok(!agents.includes('jaga:begin'))
  assert.equal(await readFile(join(root, 'src/app.js'), 'utf8'), 'console.log(1)\n')
})

test('uninstall tidak menghapus file user yang konflik', async () => {
  const root = await tempRepo(['.opencode'])
  await install(root, { pkgRoot })
  const path = join(root, '.opencode/skills/jaga/SKILL.md')
  await writeFile(path, 'versi user\n', 'utf8')
  await install(root, { pkgRoot })
  await uninstall(root)
  assert.equal(await readFile(path, 'utf8'), 'versi user\n')
})

test('uninstall pada repo kosong aman', async () => {
  const root = await tempRepo()
  const result = await uninstall(root)
  assert.equal(result.none, true)
})
