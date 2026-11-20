import { buildSection, sectionBody } from './shared.mjs'

export const target = {
  id: 'generic',
  label: 'AGENTS.md (generic)',
  async detect() {
    return true
  },
  files() {
    return [{ rel: 'AGENTS.md', content: buildSection(sectionBody), kind: 'section' }]
  },
}
