import { readdir, readFile, stat } from 'node:fs/promises'
import { join, relative } from 'node:path'

const SKIP = new Set(['.git', 'node_modules', 'dist', 'build', 'vendor', 'coverage', 'testdata'])
const RANK = { low: 1, medium: 2, high: 3, critical: 4 }

const RULES = [
  { id: 'secret', severity: 'critical', label: 'AWS access key', re: /AKIA[0-9A-Z]{16}/g },
  { id: 'secret', severity: 'critical', label: 'GitHub token', re: /gh[pousr]_[A-Za-z0-9_]{36,255}/g },
  { id: 'secret', severity: 'critical', label: 'OpenAI key', re: /sk-(proj-)?[A-Za-z0-9_-]{20,}/g },
  { id: 'secret', severity: 'critical', label: 'private key', re: /-----BEGIN (RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----/g },
  { id: 'pii', severity: 'high', label: 'NIK/KK (16 digit)', re: /\b[0-9]{16}\b/g },
  { id: 'pii', severity: 'high', label: 'nomor HP', re: /\b08[0-9]{8,11}\b/g },
]

function mask(s) {
  return s.length <= 8 ? '*'.repeat(s.length) : `${s.slice(0, 4)}…${s.slice(-4)}`
}

async function walk(root, dir, out) {
  let entries
  try {
    entries = await readdir(join(root, dir), { withFileTypes: true })
  } catch {
    return
  }
  for (const e of entries) {
    const rel = dir ? `${dir}/${e.name}` : e.name
    if (e.isDirectory()) {
      if (!SKIP.has(e.name) && !e.name.startsWith('.')) await walk(root, rel, out)
      continue
    }
    if (e.name.endsWith('_test.go')) continue
    const path = join(root, rel)
    try {
      const info = await stat(path)
      if (!info.isFile() || info.size > 2 << 20) continue
      out.push({ path, rel })
    } catch {
      // abaikan
    }
  }
}

function parseArgs(argv) {
  const opts = { format: 'text', failOn: 'high', root: '.' }
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]
    if (a === '--format' && argv[i + 1]) opts.format = argv[++i]
    else if (a === '--fail-on' && argv[i + 1]) opts.failOn = argv[++i]
    else if (!a.startsWith('-')) opts.root = a
  }
  return opts
}

async function scan(opts) {
  const files = []
  await walk(opts.root, '', files)
  const findings = []
  for (const file of files) {
    let text
    try {
      text = await readFile(file.path, 'utf8')
    } catch {
      continue
    }
    for (const rule of RULES) {
      rule.re.lastIndex = 0
      for (const m of text.matchAll(rule.re)) {
        const line = text.slice(0, m.index).split('\n').length
        findings.push({
          ruleId: rule.id,
          severity: rule.severity,
          level: RANK[rule.severity] >= 3 ? 'error' : 'warning',
          message: `${rule.label}: ${mask(m[0])}`,
          file: file.rel,
          line,
        })
      }
    }
  }
  return findings
}

function renderText(findings, root) {
  const lines = [`jaga-skill (fallback JS) — ${root}`, '─'.repeat(56)]
  if (findings.length === 0) lines.push('✓ tidak ada temuan')
  for (const f of findings) lines.push(`! ${f.ruleId.padEnd(8)} ${f.file}:${f.line} — ${f.message}`)
  lines.push('─'.repeat(56), `${findings.length} temuan`)
  return lines.join('\n') + '\n'
}

function renderSarif(findings) {
  const doc = {
    version: '2.1.0',
    $schema: 'https://json.schemastore.org/sarif-2.1.0.json',
    runs: [{
      tool: { driver: { name: 'jaga-skill (js fallback)' } },
      results: findings.map((f) => ({
        ruleId: f.ruleId,
        level: f.level,
        message: { text: f.message },
        locations: [{
          physicalLocation: {
            artifactLocation: { uri: f.file },
            region: { startLine: f.line },
          },
        }],
      })),
    }],
  }
  return JSON.stringify(doc, null, 2) + '\n'
}

const opts = parseArgs(process.argv.slice(2))
const findings = await scan(opts)
process.stdout.write(opts.format === 'sarif' ? renderSarif(findings) : renderText(findings, opts.root))
const threshold = RANK[opts.failOn] || 3
process.exitCode = findings.some((f) => RANK[f.severity] >= threshold) ? 1 : 0
