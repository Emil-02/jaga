import { exists } from './shared.mjs'

export const target = {
  id: 'cursor',
  label: 'Cursor',
  async detect(root) {
    return exists(`${root}/.cursor`)
  },
  files(ctx) {
    const body = `---\ndescription: Security & privacy gate. Check before commit or deploy.\nalwaysApply: true\n---\n\n${ctx.skill}\n`
    return [{ rel: '.cursor/rules/jaga.mdc', content: body, kind: 'skill' }]
  },
}
