# jaga-skill 🛡️

**Security & privacy gate untuk AI coding agent.** Satu perintah memasang pengawal yang membuat agent berhenti sebelum commit/deploy saat ada secret, data pribadi, atau konfigurasi berbahaya.

```sh
npx jaga-skill init
```

Kompatibel dengan **OpenCode · Claude Code · Codex · Gemini CLI · Cursor**.

## Kenapa

Skill pada umumnya hanya berisi instruksi. `jaga-skill` punya scanner nyata di belakangnya — [`jaga`](https://github.com/Emil-02/jaga) — yang menghasilkan laporan SARIF bisa dibaca agent. Pemeriksaan berjalan lokal, tanpa upload dan tanpa telemetri.

## Pemakaian

```sh
npx jaga-skill init
npx jaga-skill init --dry-run
npx jaga-skill init --with-hook
npx jaga-skill doctor
npx jaga-skill check --format sarif
npx jaga-skill uninstall
```

Dokumentasi lengkap: [`README.md`](README.md).

Dibuat dan dipelihara oleh [@Emil-02](https://github.com/Emil-02).
