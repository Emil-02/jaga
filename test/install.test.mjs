import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { install } from '../src/install.mjs'
import { readManifest } from '../src/manifest.mjs'
import { tempRepo } from './helpers.mjs'

const pkgRoot = fileURLToPath(new URL('..', import.meta.url))

test('install membuat file skill, references dan manifest', async () => {
  const root = await tempRepo(['.opencode', '.claude'], { 'AGENTS.md': '# Repo saya\n' })
  const report = await install(root, { pkgRoot })
  assert.ok(report.created.includes('.opencode/skills/jaga/SKILL.md'))
  assert.ok(report.created.includes('.claude/skills/jaga/SKILL.md'))
  const skill = await readFile(join(root, '.opencode/skills/jaga/SKILL.md'), 'utf8')
  assert.ok(skill.includes('name: jaga'))
  assert.ok(!skill.includes('{{VERSION}}'))
  const ref = await readFile(join(root, '.opencode/skills/jaga/references/rules.md'), 'utf8')
  assert.ok(ref.includes('AWS access key'))
  const manifest = await readManifest(root)
  assert.equal(manifest.version, '0.1.0')
  assert.ok(manifest.files.length >= 8)
})

test('install idempoten dan tidak menimpa file user', async () => {
  const root = await tempRepo(['.opencode'], { 'AGENTS.md': '# Asli\n' })
  await install(root, { pkgRoot })
  const first = await readManifest(root)
  const second = await install(root, { pkgRoot })
  assert.equal(second.created.length, 0)
  assert.deepEqual((await readManifest(root)).files, first.files)

  const skillPath = join(root, '.opencode/skills/jaga/SKILL.md')
  await writeFile(skillPath, 'versi user\n', 'utf8')
  const third = await install(root, { pkgRoot })
  assert.ok(third.skipped.includes('.opencode/skills/jaga/SKILL.md'))
  assert.equal(await readFile(skillPath, 'utf8'), 'versi user\n')
  const entry = (await readManifest(root)).files.find((f) => f.path.endsWith('SKILL.md'))
  assert.equal(entry.conflict, true)
})

test('force menimpa file yang berubah', async () => {
  const root = await tempRepo(['.opencode'])
  await install(root, { pkgRoot })
  await writeFile(join(root, '.opencode/skills/jaga/SKILL.md'), 'rusak\n', 'utf8')
  const report = await install(root, { pkgRoot, force: true })
  assert.ok(report.created.includes('.opencode/skills/jaga/SKILL.md'))
})

test('dry-run tidak menulis apa pun', async () => {
  const root = await tempRepo(['.opencode'])
  const report = await install(root, { pkgRoot, dryRun: true })
  assert.equal(report.dryRun, true)
  assert.ok(report.planned.length > 0)
  assert.equal(await readManifest(root), null)
})
