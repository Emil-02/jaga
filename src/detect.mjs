import { target as opencode } from './targets/opencode.mjs'
import { target as claude } from './targets/claude.mjs'
import { target as codex } from './targets/codex.mjs'
import { target as gemini } from './targets/gemini.mjs'
import { target as cursor } from './targets/cursor.mjs'
import { target as generic } from './targets/generic.mjs'

export const targets = [opencode, claude, codex, gemini, cursor]

export function getTarget(id) {
  const found = targets.find((t) => t.id === id)
  if (found) return found
  if (id === 'generic') return generic
  throw new Error(`target tidak dikenal: ${id}`)
}

export async function detectAgents(root) {
  const found = []
  for (const t of targets) {
    if (await t.detect(root)) found.push(t)
  }
  if (found.length === 0) found.push(generic)
  return found
}
