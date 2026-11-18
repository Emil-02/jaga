import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { detectAgents, getTarget } from './detect.mjs'
import { atomicWrite, hash, readManifest, writeManifest, MANIFEST_FILE } from './manifest.mjs'
import { exists, upsertSection } from './targets/shared.mjs'

const VERSION = '0.1.0'

export async function loadSkill(pkgRoot) {
  const raw = await readFile(join(pkgRoot, 'SKILL.md'), 'utf8')
  const skill = raw.replaceAll('{{VERSION}}', VERSION)
  const names = ['rules.md', 'sarif.md', 'autofix.md']
  const references = await Promise.all(
    names.map(async (name) => ({
      name,
      content: await readFile(join(pkgRoot, 'references', name), 'utf8'),
    })),
  )
  return { skill, references }
}

export async function install(root, opts = {}) {
  const pkgRoot = opts.pkgRoot || fileURLToPath(new URL('..', import.meta.url))
  const { skill, references } = await loadSkill(pkgRoot)
  const ctx = { skill, references, version: VERSION }
  const wanted = opts.targets?.length ? opts.targets.map(getTarget) : await detectAgents(root)
  const planned = wanted.flatMap((t) =>
    t.files(ctx).map((f) => ({ ...f, target: t.id, path: `${root}/${f.rel}` })),
  )
  const report = { version: VERSION, targets: wanted.map((t) => t.id), created: [], unchanged: [], skipped: [], sections: [], planned: planned.map((f) => f.rel) }

  if (opts.dryRun) {
    report.dryRun = true
    return report
  }

  const previous = await readManifest(root)
  const previousByPath = new Map((previous?.files || []).map((f) => [f.path, f]))
  const entries = []
  for (const file of planned) {
    const present = await exists(file.path)
    if (file.kind === 'section') {
      const original = present ? await readFile(file.path, 'utf8') : ''
      const updated = upsertSection(original, file.content)
      if (updated === original) {
        report.unchanged.push(file.rel)
      } else {
        await atomicWrite(file.path, updated)
        report.sections.push(file.rel)
      }
      entries.push({ path: file.rel, kind: 'section', hash: hash(updated), created: previousByPath.get(file.rel)?.created ?? !present })
      continue
    }

    if (present && !opts.force) {
      const current = await readFile(file.path, 'utf8')
      if (hash(current) !== hash(file.content)) {
        report.skipped.push(file.rel)
        entries.push({ path: file.rel, kind: 'skill', hash: hash(current), created: previousByPath.get(file.rel)?.created ?? false, conflict: true })
        continue
      }
      report.unchanged.push(file.rel)
    } else {
      await atomicWrite(file.path, file.content)
      report.created.push(file.rel)
    }
    entries.push({ path: file.rel, kind: 'skill', hash: hash(file.content), created: previousByPath.get(file.rel)?.created ?? !present })
  }

  if (previous && previous.files) {
    for (const old of previous.files) {
      if (!entries.some((e) => e.path === old.path)) entries.push(old)
    }
  }
  await writeManifest(root, { version: VERSION, targets: report.targets, files: entries })
  return report
}

export { MANIFEST_FILE }
