import { readFile, rm, stat } from 'node:fs/promises'
import { join } from 'node:path'
import { readManifest, writeManifest, atomicWrite } from './manifest.mjs'
import { exists, stripSection } from './targets/shared.mjs'

export async function uninstall(root) {
  const manifest = await readManifest(root)
  if (!manifest || !manifest.files) return { removed: [], kept: [], none: true }

  const removed = []
  const kept = []
  for (const entry of manifest.files) {
    const path = join(root, entry.path)
    if (!(await exists(path))) continue
    if (entry.kind === 'section') {
      const original = await readFile(path, 'utf8')
      const cleaned = stripSection(original)
      if (cleaned.trim() === '' && entry.created) {
        await rm(path)
        removed.push(entry.path)
      } else if (cleaned !== original) {
        await atomicWrite(path, cleaned)
        removed.push(entry.path)
      } else {
        kept.push(entry.path)
      }
      continue
    }
    if (entry.conflict) {
      kept.push(entry.path)
      continue
    }
    await rm(path)
    removed.push(entry.path)
  }

  await rm(join(root, '.jaga-skill.json'), { force: true })
  return { removed, kept }
}

export async function isGitRepo(root) {
  try {
    return (await stat(join(root, '.git'))).isDirectory()
  } catch {
    return false
  }
}
