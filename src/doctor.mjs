import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { readManifest } from './manifest.mjs'
import { resolveBinary } from './binary.mjs'
import { exists } from './targets/shared.mjs'
import { isGitRepo } from './uninstall.mjs'

export async function doctor(root, opts = {}) {
  const pkgRoot = opts.pkgRoot || fileURLToPath(new URL('..', import.meta.url))
  const checks = []
  const nodeOk = Number(process.versions.node.split('.')[0]) >= 20
  checks.push({ label: 'Node.js >= 20', ok: nodeOk, detail: process.versions.node })

  checks.push({
    label: 'SKILL.md tersedia',
    ok: await exists(join(pkgRoot, 'SKILL.md')),
    detail: 'paket inti',
  })

  const manifest = await readManifest(root)
  checks.push({
    label: '.jaga-skill.json',
    ok: !!manifest,
    detail: manifest ? `${manifest.files?.length || 0} file dikelola` : 'belum dijalankan init',
  })

  if (manifest?.files) {
    let bad = 0
    for (const entry of manifest.files) {
      if (!(await exists(join(root, entry.path)))) {
        bad++
        continue
      }
      if (entry.kind === 'skill') {
        const text = await readFile(join(root, entry.path), 'utf8')
        const { hash } = await import('./manifest.mjs')
        if (hash(text) !== entry.hash) bad++
      }
    }
    checks.push({ label: 'Integritas file', ok: bad === 0, detail: bad ? `${bad} file berubah` : 'semua sesuai manifest' })
  }

  const bin = await resolveBinary()
  checks.push({ label: 'Scanner', ok: true, detail: `${bin.kind}: ${bin.note}` })
  checks.push({ label: 'Git repository', ok: await isGitRepo(root), detail: root })

  return checks
}
