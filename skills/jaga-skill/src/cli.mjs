#!/usr/bin/env node
import { spawnSync } from 'node:child_process'
import { install } from './install.mjs'
import { uninstall, isGitRepo } from './uninstall.mjs'
import { doctor } from './doctor.mjs'
import { checkCommand } from './binary.mjs'
import { readConfig } from './config.mjs'
import { installHook, uninstallHook } from './hook.mjs'

const VERSION = '0.1.0'

function usage() {
  process.stdout.write(`jaga-skill ${VERSION} — security gate untuk AI coding agent

Pemakaian:
  jaga-skill init [--root DIR] [--dry-run] [--force] [--with-hook] [--target ID]
  jaga-skill check [--root DIR] [--format text|sarif] [--fail-on LEVEL]
  jaga-skill doctor [--root DIR]
  jaga-skill uninstall [--root DIR] [--yes]
  jaga-skill update [--root DIR]
  jaga-skill version
`)
}

function parse(argv) {
  const flags = { root: '.', targets: [], passthrough: [] }
  let rest = false
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]
    if (rest) {
      flags.passthrough.push(a)
      continue
    }
    switch (a) {
      case '--':
        rest = true
        break
      case '--root':
        flags.root = argv[++i]
        break
      case '--target':
        flags.targets.push(argv[++i])
        break
      case '--format':
      case '--fail-on':
        flags[a.slice(2)] = argv[++i]
        break
      case '--dry-run':
        flags.dryRun = true
        break
      case '--force':
        flags.force = true
        break
      case '--with-hook':
        flags.withHook = true
        break
      case '--yes':
        flags.yes = true
        break
      default:
        if (a.startsWith('-')) throw new Error(`opsi tidak dikenal: ${a}`)
        flags.root = a
    }
  }
  return flags
}

async function main() {
  const [cmd, ...argv] = process.argv.slice(2)
  if (!cmd || cmd === 'help' || cmd === '-h' || cmd === '--help') return usage()
  if (cmd === 'version') return process.stdout.write(`jaga-skill ${VERSION}\n`)

  const flags = parse(argv)
  const root = flags.root

  if (cmd === 'init' || cmd === 'update') {
    const report = await install(root, {
      dryRun: flags.dryRun,
      force: flags.force || cmd === 'update',
      targets: flags.targets,
    })
    if (report.dryRun) {
      process.stdout.write(`Rencana (${report.targets.join(', ')}):\n${report.planned.map((p) => `  ${p}`).join('\n')}\n`)
      return
    }
    process.stdout.write(`Target: ${report.targets.join(', ')}\n`)
    process.stdout.write(`Dibuat: ${report.created.length}, diubah: ${report.sections.length}, tidak berubah: ${report.unchanged.length}, dilewati: ${report.skipped.length}\n`)
    for (const p of report.skipped) process.stdout.write(`  skipped (file berbeda): ${p}\n`)
    if (flags.withHook) {
      const hook = await installHook(root)
      process.stdout.write(hook.ok ? `Hook: ${hook.path}\n` : `Hook: gagal — ${hook.reason}\n`)
    }
    return
  }

  if (cmd === 'uninstall') {
    if (!(await isGitRepo(root)) && !flags.yes) {
      throw new Error('root bukan repo git; gunakan --yes untuk tetap melanjutkan')
    }
    const hook = await uninstallHook(root)
    const result = await uninstall(root)
    if (result.none) return process.stdout.write('Tidak ada instalasi ditemukan.\n')
    process.stdout.write(`Dihapus: ${result.removed.length}, dipertahankan: ${result.kept.length}${hook.removed ? ', hook dilepas' : ''}\n`)
    return
  }

  if (cmd === 'doctor') {
    const checks = await doctor(root)
    for (const c of checks) {
      process.stdout.write(`${c.ok ? '✓' : '✗'} ${c.label} — ${c.detail}\n`)
    }
    process.exitCode = checks.every((c) => c.ok) ? 0 : 1
    return
  }

  if (cmd === 'check') {
    const bin = await checkCommand()
    const cfg = await readConfig(root)
    const args = [...bin.prefixArgs]
    if (bin.kind !== 'js') args.push('check')
    args.push('--format', flags.format || 'text', '--fail-on', flags.failOn || cfg.fail_on)
    args.push(...flags.passthrough, root)
    const res = spawnSync(bin.bin, args, { stdio: 'inherit' })
    if (res.error) throw res.error
    process.exitCode = res.status ?? 1
    return
  }

  throw new Error(`perintah tidak dikenal: ${cmd}`)
}

try {
  await main()
} catch (err) {
  process.stderr.write(`jaga-skill: ${err.message}\n`)
  process.exitCode = 2
}
