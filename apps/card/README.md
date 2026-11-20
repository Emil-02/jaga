# Jaga Card

Kartu skor kepercayaan untuk repo GitHub yang diamankan `jaga`.

Aplikasi ini adalah bagian dari [repo jaga](../../README.md), berada di `apps/card`. Tidak perlu repo terpisah.

Jaga Card membaca `jaga-report.json` dari repo, lalu merangkum skor secret, data pribadi, dan hygiene menjadi satu halaman + badge SVG. Tanpa database, tanpa telemetri, tanpa scan server-side.

## Cara pakai

Buka `/` dan masukkan either:

- `Emil-02` → halaman profil dan rata-rata skor semua repo
- `Emil-02/jaga` → halaman detail satu repo

Atau langsung buka `/u/<username>` / `/r/<owner>/<repo>`.

## Memasang laporan di repo

Jalankan di repo yang ingin ditampilkan:

```sh
jaga check . --format json -o jaga-report.json
git add jaga-report.json
git commit -m "chore: add jaga report"
git push
```

Jaga Card membaca file dari branch default melalui `raw.githubusercontent.com`, jadi tidak perlu token untuk repo publik.

### Otomatis lewat GitHub Actions

Tambahkan `.github/workflows/jaga.yml`. Supaya commit tetap atas nama kamu (bukan bot), pakai PAT pribadi sebagai `JAGA_PAT`:

```yaml
name: jaga report
on:
  push:
    branches: [main]
  schedule:
    - cron: "17 3 * * 1"

permissions:
  contents: write

jobs:
  report:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          token: ${{ secrets.JAGA_PAT }}
      - uses: Emil-02/jaga@main
        with:
          report: "true"
          badge: "true"
      - name: Commit laporan
        run: |
          git config user.name "emil rahmat"
          git config user.email "189325051+Emil-02@users.noreply.github.com"
          git add jaga-report.json badge.svg
          if ! git diff --cached --quiet; then
            git commit -m "chore(jaga): update report"
            git push
          fi
```

Nilai `user.name` dan `user.email` di atas memastikan author & committer tetap kamu. Jangan memakai `github-actions[bot]` kalau kamu ingin satu-satunya contributor tetap akun kamu.

## API

| Endpoint | Keterangan |
|---|---|
| `GET /api/badge?owner=…&repo=…` | SVG badge skor satu repo |
| `GET /api/badge?user=…` | SVG badge rata-rata skor profil |
| `GET /api/card?user=…` | SVG kartu profil 560×150 |

Contoh di README:

```md
[![jaga](https://jaga-card.vercel.app/api/badge?owner=Emil-02&repo=jaga)](https://jaga-card.vercel.app/r/Emil-02/jaga)
```

Ganti domain dengan URL deploy Vercel kamu bila berbeda.

## Environment

- `GITHUB_TOKEN` (opsional) — token GitHub untuk menaikkan batas rate limit API (60 → 5000 request/jam). Simpan sebagai environment variable di Vercel, jangan di-commit.

Tanpa token, halaman profil membatasi 24 repo dan cache respons 10 menit.

## Development

Dari root repo:

```sh
cd apps/card
npm install
npm run dev
npm run lint
npm run typecheck
npm run build
```

## Deploy

Di Vercel, arahkan **Root Directory** ke `apps/card`. Environment variable opsional: `GITHUB_TOKEN`.

## Lisensi

MIT, bagian dari repo `jaga`. Dibuat dan dipelihara oleh [@Emil-02](https://github.com/Emil-02).
