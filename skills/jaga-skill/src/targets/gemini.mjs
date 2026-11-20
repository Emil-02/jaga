import { exists, buildSection, sectionBody } from './shared.mjs'

export const target = {
  id: 'gemini',
  label: 'Gemini CLI',
  async detect(root) {
    return (await exists(`${root}/.gemini`)) || (await exists(`${root}/GEMINI.md`))
  },
  files() {
    return [{ rel: 'GEMINI.md', content: buildSection(sectionBody), kind: 'section' }]
  },
}
