import { chmod, readFile, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { atomicWrite } from './manifest.mjs'
import { exists } from './targets/shared.mjs'

const BEGIN = '# jaga:begin'
const END = '# jaga:end'
const BODY = `jaga-skill check || {
  echo "jaga: commit diblokir oleh pemeriksaan keamanan" >&2
  exit 1
}`

const block = `${BEGIN}\n${BODY}\n${END}`

function strip(text) {
  const a = text.indexOf(BEGIN)
  const b = text.indexOf(END)
  if (a < 0 || b < a) return text
  const before = text.slice(0, a).trimEnd()
  const after = text.slice(b + END.length).trimStart()
  return [before, after].filter(Boolean).join('\n\n') + (before || after ? '\n' : '')
}

export async function installHook(root) {
  const path = join(root, '.git', 'hooks', 'pre-commit')
  if (!(await exists(join(root, '.git')))) return { ok: false, reason: 'bukan repo git' }
  const original = (await exists(path)) ? await readFile(path, 'utf8') : '#!/bin/sh\n'
  const cleaned = strip(original)
  const next = `${cleaned.trimEnd()}\n${block}\n`
  await atomicWrite(path, next)
  await chmod(path, 0o755)
  return { ok: true, path: '.git/hooks/pre-commit' }
}

export async function uninstallHook(root) {
  const path = join(root, '.git', 'hooks', 'pre-commit')
  if (!(await exists(path))) return { ok: true, removed: false }
  const original = await readFile(path, 'utf8')
  const cleaned = strip(original)
  if (cleaned === original) return { ok: true, removed: false }
  await writeFile(path, cleaned, 'utf8')
  return { ok: true, removed: true }
}
