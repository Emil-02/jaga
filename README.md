# jaga 🌿

**Satu perintah untuk memeriksa keamanan, privasi, dan kesehatan repo sebelum kode kamu jadi publik.**

`jaga` menyatukan pemeriksaan yang selama ini tersebar di tool-tool saya menjadi satu gerbang: satu binary Go, tanpa server, tanpa akun, bisa dipakai lokal maupun sebagai GitHub Action.

> **Pakai AI coding agent?** Pasang [`jaga-skill`](skills/jaga-skill/README.md) dengan `npx jaga-skill init` — gate yang menghentikan agent sebelum commit/deploy saat ada secret atau data pribadi. `jaga` adalah scanner di balik skill tersebut.

```console
$ jaga check .

jaga v0.1.0 — /home/me/project
────────────────────────────────────────────────────────
✓ secret       tidak ada kredensial terekspos
! pii          2 data pribadi ditemukan di README.md
✓ hygiene      LICENSE, README, CI, .gitignore lengkap
────────────────────────────────────────────────────────
Skor: 88/100      1 peringatan      0 error
```

## Kenapa jaga?

Saya punya beberapa tool open source dengan tema yang sama: [`envboy`](https://github.com/Emil-02/envboy) untuk secret, [`sensor`](https://github.com/Emil-02/sensor) untuk data pribadi, [`leakboy`](https://github.com/Emil-02/leakboy) untuk kebocoran di aplikasi yang sudah deploy, [`gh-spruce`](https://github.com/Emil-02/gh-spruce) untuk kesehatan repo, dan [`cekdulu`](https://github.com/Emil-02/cekdulu) untuk deteksi penipuan. `jaga` adalah pintu masuknya: satu tempat yang menjalankannya sebelum kamu push atau deploy.

- **Developer:** satu perintah, output terminal yang enak dibaca, JSON untuk diproses, SARIF untuk GitHub Code Scanning, dan mode `--ci` yang gagal bila ada temuan berat.
- **Masyarakat:** aplikasi yang dibangun developer jadi lebih kecil kemungkinannya membocorkan NIK, rekening, atau API key.

## Fitur

- Pemeriksaan **secret AI & Cloud modern** (OpenAI, Anthropic, Google Gemini, Groq, HuggingFace, Supabase `service_role` vs `anon`, Database URI, Midtrans, Xendit, Stripe, AWS, GitHub) dengan masking otomatis dan filter placeholder.
- Pemeriksaan **data pribadi Indonesia** (NIK, NPWP, nomor HP, rekening) — logika dari [`sensor`](https://github.com/Emil-02/sensor).
- Pemeriksaan **kebocoran frontend vibe-coding** (deteksi variabel `NEXT_PUBLIC_*` / `VITE_*` yang berisi private/service keys).
- Pemeriksaan **hygiene repo** (LICENSE, README, workflow CI, .gitignore).
- Konfigurasi `.jaga.yml`: exclude path, atur severity, atur ambang gagal.
- Report: `text`, `json`, `sarif`, `badge` (SVG), `html` (dashboard mandiri), dan `markdown` (untuk GitHub Step Summary & PR).
- `jaga-mcp`: MCP server stdio untuk AI coding agent (Claude Code, Codex, Gemini CLI, OpenCode, Cursor).
- Berjalan paralel, tanpa mengirim data keluar (semua lokal).
- [Panduan Menambahkan Rule Baru](docs/adding-rules.md) untuk kontributor open source.

## Instalasi

```sh
go install github.com/Emil-02/jaga/cmd/jaga@latest
go install github.com/Emil-02/jaga/cmd/jaga-mcp@latest
```

Atau build dari sumber:

```sh
git clone https://github.com/Emil-02/jaga.git
cd jaga
go build -o jaga ./cmd/jaga
```

## Pemakaian

```sh
jaga check .                       # periksa direktori saat ini
jaga check ./app --format json     # output mesin
jaga check . --format sarif -o jaga.sarif
jaga check . --format badge -o badge.svg
jaga check . --format html -o laporan.html   # dashboard, tanpa JS/CDN
jaga check . --ci --fail-on high   # untuk pipeline
```

Keluar dengan kode `1` bila ada temuan pada/melebihi `--fail-on` (default: `high`).

## Menyiapkan repo baru

```sh
jaga init . \
  --author-name "emil rahmat" \
  --author-email "189325051+Emil-02@users.noreply.github.com"
```

`jaga init` menulis `.jaga.yml`, workflow `.github/workflows/jaga.yml` (PR Gate + generator laporan),
badge di README, lalu menjalankan scan pertama menjadi `jaga-report.json`. File yang sudah ada tidak
ditimpa kecuali pakai `--force`; identitas commit di workflow diambil dari argumen atau environment
`JAGA_AUTHOR_NAME` / `JAGA_AUTHOR_EMAIL` — tidak memakai `github-actions[bot]`.

Workflow commit laporan memakai PAT pribadi (`secrets.JAGA_PAT`) supaya author/committer tetap akun kamu.
Tanpa PAT, laporan tetap dibuat sebagai artifact tapi tidak di-commit.

## Perbaikan otomatis

```sh
jaga fix .                 # dry-run: tampilkan rencana
jaga fix . --format patch  # bentuk diff
jaga fix . --apply         # tulis hanya perbaikan aman
jaga fix . --report jaga-report.json --apply
```

Perbaikan **aman** hanya menyentuh file konfigurasi: membuat `LICENSE`, README dasar, workflow CI,
dan menambah entri `.gitignore` untuk `.env`/kunci. File yang sudah ada tidak pernah ditimpa.
Temuan `secret` dan `pii` **tidak pernah diubah otomatis** — hanya muncul sebagai saran rotasi/masking.
Tool MCP `jaga_fix` memakai default dry-run; agent harus menyetel `apply: true` secara eksplisit.

## Konfigurasi

Buat `.jaga.yml` di root repo (lihat [`examples/.jaga.yml`](examples/.jaga.yml)):

```yaml
version: 1
exclude:
  - node_modules
  - dist
  - vendor
checks:
  secret:
    severity: critical
  pii:
    severity: high
  hygiene:
    severity: low
fail_on: high
```

## GitHub Action

```yaml
- uses: Emil-02/jaga@main
  with:
    path: .
    fail-on: high
    summary: "true" # Tulis tabel ringkasan ke GitHub Step Summary
    sarif: "true"   # Tulis jaga.sarif untuk GitHub Code Scanning
```

Set `report: "true"` dan `badge: "true"` untuk menghasilkan `jaga-report.json` dan `badge.svg`. Commit file laporan ke repo supaya [Jaga Card](apps/card/README.md) bisa menampilkan skor kamu sebagai kartu publik.

### Contoh Workflow Lengkap dengan GitHub Code Scanning:

```yaml
name: Security & Health Check

on:
  push:
    branches: [main]
  pull_request:

jobs:
  jaga:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - name: Run Jaga Gate
        uses: Emil-02/jaga@main
        with:
          fail-on: high
          sarif: "true"
          summary: "true"
      - name: Upload SARIF report
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: jaga.sarif
```

## PR Gate

`jaga-diff` membandingkan dua laporan, lalu workflow [`jaga-pr.yml`](.github/workflows/jaga-pr.yml) memposting ringkasan `Skor 88 → 76`, temuan baru, dan temuan selesai ke PR. Opsi `strict: true` menggagalkan workflow bila skor turun atau ada temuan baru pada/melebihi `fail-on`.

Pakai di repo lain:

```yaml
# .github/workflows/jaga.yml
name: jaga PR Gate
on: pull_request
permissions:
  contents: read
  pull-requests: write
jobs:
  gate:
    uses: Emil-02/jaga/.github/workflows/jaga-pr.yml@main
    with:
      strict: true
      fail-on: high
```

Komentar dibuat oleh `github-actions[bot]` lewat `GITHUB_TOKEN` — hanya komentar, bukan commit, jadi tidak menambah contributor ke repo.

Secara lokal:

```sh
jaga check /path/ke/base --format json -o base.json
jaga check . --format json -o head.json
jaga-diff base.json head.json --format markdown
```

## MCP server untuk AI agent

`jaga-mcp` membuka scanner yang sama lewat MCP over stdio, supaya agent bisa memeriksa kode
**sebelum commit atau deploy**. Masih stdlib Go, tanpa dependensi eksternal.

Tool yang tersedia:

| Tool | Fungsi |
|---|---|
| `jaga_check` | Skor 0–100 + grade + temuan (path, baris, pesan yang sudah di-masking). Bisa dibatasi ke `secret`, `pii`, atau `hygiene`. |
| `jaga_gate` | Gerbang commit/deploy: `allowed: true` hanya bila tidak ada temuan pada/melebihi `fail_on`. |
| `jaga_badge` | SVG badge siap tempel di README. |
| `jaga_fix` | Rencana perbaikan (dry-run default); `apply: true` hanya menulis file konfigurasi aman. |

Server hanya boleh memindai direktori di bawah `-root` (default: direktori kerja). Path di luar root
ditolak, dan semua temuan dibatasi maksimal 100 item per respons.

**OpenCode** — tambahkan ke `opencode.jsonc` di project (V2 memakai `mcp.servers`):

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "servers": {
      "jaga": {
        "type": "local",
        "command": ["jaga-mcp", "-root", "/path/ke/project"],
        "codemode": false
      }
    }
  }
}
```

**Agent lain** (Claude Code, Cursor, Codex, Gemini CLI) — pola umum `mcpServers`:

```json
{
  "mcpServers": {
    "jaga": {
      "command": "jaga-mcp",
      "args": ["-root", "/path/ke/project"]
    }
  }
}
```

Cek koneksi dengan `opencode mcp list`, lalu ajakan agent: *"jalankan jaga_gate sebelum commit"*.

## Dashboard HTML & badge

```sh
jaga check . --format html -o laporan.html
jaga check . --format badge -o badge.svg
```

`laporan.html` adalah satu file mandiri: ring skor, ringkasan temuan, dan kartu per pemeriksaan.
Tidak ada JavaScript, font, atau CSS dari CDN — aman dibuka offline dan bisa di-commit ke repo
atau di-host sebagai artifact CI.

## Arsitektur

```
cmd/jaga        entrypoint CLI (stdlib flag, tanpa dependensi eksternal)
cmd/jaga-mcp    MCP server stdio (JSON-RPC 2.0, stdlib)
cmd/jaga-diff   pembanding dua laporan untuk PR Gate
internal/setup  generator file saat `jaga init`
internal/fix    perencana & penulis perbaikan aman
apps/card       Jaga Card — web Next.js (Vercel, root directory: apps/card)
skills/jaga-skill   paket npm untuk gate AI coding agent
internal/check  implementasi pemeriksaan + interface umum
internal/runner eksekusi paralel + timeout
internal/config parser .jaga.yml (subset yang didokumentasikan)
internal/report formatter text/json/sarif/badge/html
internal/score  perhitungan skor 0-100
```

Rencana lanjutan: integrasi penuh `leakboy` dan `gh-spruce`, rule packs dari komunitas, serta dukungan konfigurasi YAML penuh.

## Rilis

Satu repo, dua artefact dengan tag berbeda:

- `v*` — rilis `jaga` (CLI, MCP, Action) lewat `go install`.
- `skill-v*` — publish `jaga-skill` ke npm lewat `.github/workflows/skill-release.yml`.

## Lisensi & atribusi

MIT. Dibuat dan dipelihara oleh [@Emil-02](https://github.com/Emil-02). Seluruh kepemilikan dan atribusi kode di repository ini mengacu pada akun GitHub tersebut (lihat [`.github/CODEOWNERS`](.github/CODEOWNERS)).
