import { access } from 'node:fs/promises'

export async function exists(p) {
  try {
    await access(p)
    return true
  } catch {
    return false
  }
}

export const BEGIN = '<!-- jaga:begin -->'
export const END = '<!-- jaga:end -->'

export function buildSection(body) {
  return `${BEGIN}\n${body.trim()}\n${END}`
}

export function upsertSection(original, block) {
  const start = original.indexOf(BEGIN)
  const end = original.indexOf(END)
  if (start >= 0 && end > start) {
    const before = original.slice(0, start).trimEnd()
    const after = original.slice(end + END.length).trimStart()
    return [before, block, after].filter(Boolean).join('\n\n') + '\n'
  }
  const prefix = original.trimEnd()
  return prefix ? `${prefix}\n\n${block}\n` : `${block}\n`
}

export function stripSection(original) {
  const start = original.indexOf(BEGIN)
  const end = original.indexOf(END)
  if (start < 0 || end < start) return original
  const before = original.slice(0, start).trimEnd()
  const after = original.slice(end + END.length).trimStart()
  return [before, after].filter(Boolean).join('\n\n') + (before || after ? '\n' : '')
}

export function skillFiles(skillDir, ctx) {
  const files = [{ rel: `${skillDir}/SKILL.md`, content: ctx.skill, kind: 'skill' }]
  for (const ref of ctx.references) {
    files.push({
      rel: `${skillDir}/references/${ref.name}`,
      content: ref.content,
      kind: 'skill',
    })
  }
  return files
}

export const sectionBody = `
## jaga — security gate

File ini dikelola oleh \`jaga-skill\`. Sebelum commit atau deploy, jalankan \`jaga-skill check\`.
Jika laporan SARIF berisi \`level: error\`, hentikan pekerjaan dan perbaiki temuan terlebih dahulu.
Jangan pernah mencetak secret atau data pribadi mentah, dan jangan auto-push tanpa konfirmasi.
Lihat \`.jaga-skill.json\` untuk daftar file yang dikelola.
`
