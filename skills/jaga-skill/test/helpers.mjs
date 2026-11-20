import { mkdtemp, mkdir, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

export async function tempRepo(dirs = [], files = {}) {
  const root = await mkdtemp(join(tmpdir(), 'jaga-skill-test-'))
  for (const d of dirs) await mkdir(join(root, d), { recursive: true })
  for (const [rel, content] of Object.entries(files)) {
    await mkdir(join(root, rel.split('/').slice(0, -1).join('/')), { recursive: true })
    await writeFile(join(root, rel), content, 'utf8')
  }
  return root
}
