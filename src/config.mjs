import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { atomicWrite } from './manifest.mjs'

export const CONFIG_FILE = 'jaga.skill.yml'

export const defaultConfig = {
  fail_on: 'high',
  exclude: [],
  checks: { secret: true, pii: true, hygiene: true },
}

export async function readConfig(root) {
  let raw
  try {
    raw = await readFile(join(root, CONFIG_FILE), 'utf8')
  } catch {
    return { ...defaultConfig }
  }
  const cfg = { ...defaultConfig, checks: { ...defaultConfig.checks } }
  for (const line of raw.split(/\r?\n/)) {
    const t = line.trim()
    if (!t || t.startsWith('#')) continue
    const m = t.match(/^([\w_]+):\s*(.*)$/)
    if (!m) continue
    if (m[1] === 'fail_on') cfg.fail_on = m[2].trim()
    if (m[1] === 'checks') {
      const sub = line.match(/^\s+([\w_]+):\s*(true|false)$/)
      if (sub) cfg.checks[sub[1]] = sub[2] === 'true'
    }
  }
  return cfg
}

export async function writeConfig(root, cfg) {
  const body = `fail_on: ${cfg.fail_on}\nexclude:\n${cfg.exclude.map((x) => `  - ${x}`).join('\n')}\nchecks:\n${Object.entries(cfg.checks).map(([k, v]) => `  ${k}: ${v}`).join('\n')}\n`
  await atomicWrite(join(root, CONFIG_FILE), body)
}
