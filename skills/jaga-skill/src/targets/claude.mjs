import { exists, skillFiles } from './shared.mjs'

export const target = {
  id: 'claude',
  label: 'Claude Code',
  async detect(root) {
    return exists(`${root}/.claude`)
  },
  files(ctx) {
    return skillFiles('.claude/skills/jaga', ctx)
  },
}
