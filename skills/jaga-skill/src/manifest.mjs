import { mkdir, readFile, rename, writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { createHash } from 'node:crypto'

export const MANIFEST_FILE = '.jaga-skill.json'

export function hash(text) {
  return 'sha256:' + createHash('sha256').update(text).digest('hex')
}

export async function atomicWrite(file, content) {
  await mkdir(dirname(file), { recursive: true })
  const tmp = `${file}.tmp-${process.pid}`
  await writeFile(tmp, content, 'utf8')
  await rename(tmp, file)
}

export async function readManifest(root) {
  try {
    const raw = await readFile(join(root, MANIFEST_FILE), 'utf8')
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export async function writeManifest(root, manifest) {
  await atomicWrite(join(root, MANIFEST_FILE), JSON.stringify(manifest, null, 2) + '\n')
}
