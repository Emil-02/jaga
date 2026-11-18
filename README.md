# jaga-skill 🛡️

**Security & privacy gate untuk AI coding agent.** Satu perintah memasang pengawal yang membuat agent berhenti sebelum commit/deploy saat ada secret, data pribadi, atau konfigurasi berbahaya.

```sh
npx jaga-skill init
```

Kompatibel dengan **OpenCode · Claude Code · Codex · Gemini CLI · Cursor**, plus fallback `AGENTS.md` untuk agent lain.

## Kenapa ini beda

Skill kebanyakan hanya berisi instruksi "hati-hati ya". `jaga-skill` punya scanner deterministik di belakangnya: [`jaga`](https://github.com/Emil-02/jaga) (Go, read-only, lokal) menghasilkan SARIF yang bisa dibaca agent dan diperbaiki sendiri. Tidak ada upload, tidak ada telemetri.

## Alur kerja

```
npx jaga-skill init → deteksi agent → tulis SKILL.md + references + manifest
        ↓
agent mau commit/deploy → jaga-skill check → SARIF → perbaiki → check lagi
```

## Pemakaian

```sh
npx jaga-skill init              # pasang ke semua agent yang terdeteksi
npx jaga-skill init --dry-run    # lihat rencana tanpa menulis
npx jaga-skill init --with-hook  # + git pre-commit hook
npx jaga-skill doctor            # cek kesehatan instalasi
npx jaga-skill check             # jalankan pemeriksaan
npx jaga-skill check --format sarif
npx jaga-skill uninstall         # hapus hanya file yang dikelola
```

File yang ditulis:

| Agent | Lokasi |
|---|---|
| OpenCode | `.opencode/skills/jaga/SKILL.md` |
| Claude Code | `.claude/skills/jaga/SKILL.md` |
| Codex | `AGENTS.md` (blok bermarker) |
| Gemini CLI | `GEMINI.md` (blok bermarker) |
| Cursor | `.cursor/rules/jaga.mdc` |
| Fallback | `AGENTS.md` |

Semua tercatat di `.jaga-skill.json` (hash + versi) supaya uninstall aman dan install ulang idempoten.

## Resolver binary

`jaga-skill check` memilih, berurutan:
1. `JAGA_BIN` / `jaga` yang sudah ada di PATH (scanner penuh)
2. scanner JS bawaan (subset aturan, tanpa jaringan, langsung dipakai)
3. Go via `JAGA_USE_GO=1` (opt-in, `go run github.com/Emil-02/jaga/cmd/jaga@latest`)
4. binary rilis dengan verifikasi SHA-256 (rencana)

## Konfigurasi

`jaga.skill.yml` (opsional):

```yaml
fail_on: high
exclude:
  - dist
  - node_modules
checks:
  secret: true
  pii: true
  hygiene: true
```

Schema: [`schema/config.schema.json`](schema/config.schema.json).

## Struktur

Lihat [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Lisensi & atribusi

MIT. Dibuat dan dipelihara oleh [@Emil-02](https://github.com/Emil-02). Seluruh atribusi dan kepemilikan repository ini mengacu pada akun GitHub tersebut (lihat [`.github/CODEOWNERS`](.github/CODEOWNERS)); tidak ada co-author atau kontributor lain.
