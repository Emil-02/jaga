import { exists, buildSection, sectionBody } from './shared.mjs'

export const target = {
  id: 'codex',
  label: 'Codex',
  async detect(root) {
    return exists(`${root}/.codex`)
  },
  files() {
    return [{ rel: 'AGENTS.md', content: buildSection(sectionBody), kind: 'section' }]
  },
}
