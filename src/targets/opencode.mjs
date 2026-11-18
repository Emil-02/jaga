import { exists, skillFiles } from './shared.mjs'

export const target = {
  id: 'opencode',
  label: 'OpenCode',
  async detect(root) {
    return exists(`${root}/.opencode`)
  },
  files(ctx) {
    return skillFiles('.opencode/skills/jaga', ctx)
  },
}
