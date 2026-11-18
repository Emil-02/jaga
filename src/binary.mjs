import { access, constants } from 'node:fs/promises'
import { delimiter, join } from 'node:path'
import { fileURLToPath } from 'node:url'

export const JS_SCANNER = fileURLToPath(new URL('./js-scanner.mjs', import.meta.url))

async function which(name) {
  const paths = (process.env.PATH || '').split(delimiter).filter(Boolean)
  for (const dir of paths) {
    const candidate = join(dir, name)
    try {
      await access(candidate, constants.X_OK)
      return candidate
    } catch {
      // lanjut
    }
  }
  return null
}

export async function resolveBinary() {
  if (process.env.JAGA_BIN) {
    return { kind: 'env', bin: process.env.JAGA_BIN, prefixArgs: [], note: 'JAGA_BIN' }
  }
  const jaga = await which(process.platform === 'win32' ? 'jaga.exe' : 'jaga')
  if (jaga) return { kind: 'path', bin: jaga, prefixArgs: [], note: jaga }
  if (process.env.JAGA_USE_GO === '1') {
    const go = await which(process.platform === 'win32' ? 'go.exe' : 'go')
    if (go) {
      return {
        kind: 'go',
        bin: go,
        prefixArgs: ['run', 'github.com/Emil-02/jaga/cmd/jaga@latest'],
        note: 'go run (opt-in JAGA_USE_GO=1, butuh internet)',
      }
    }
  }
  return {
    kind: 'js',
    bin: process.execPath,
    prefixArgs: [JS_SCANNER],
    note: 'fallback scanner JS (subset aturan)',
  }
}

export async function checkCommand() {
  return resolveBinary()
}
