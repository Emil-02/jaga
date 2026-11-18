# jaga 🌿

**Satu perintah untuk memeriksa keamanan, privasi, dan kesehatan repo sebelum kode kamu jadi publik.**

`jaga` menyatukan pemeriksaan yang selama ini tersebar di tool-tool saya menjadi satu gerbang: satu binary Go, tanpa server, tanpa akun, bisa dipakai lokal maupun sebagai GitHub Action.

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

- Pemeriksaan **secret** (AWS, GitHub, OpenAI, private key, API key generik) dengan masking otomatis.
- Pemeriksaan **data pribadi Indonesia** (NIK, NPWP, nomor HP, rekening) — logika dari [`sensor`](https://github.com/Emil-02/sensor).
- Pemeriksaan **hygiene repo** (LICENSE, README, workflow CI, .gitignore).
- Konfigurasi `.jaga.yml`: exclude path, atur severity, atur ambang gagal.
- Report: `text`, `json`, `sarif`, `badge` (SVG).
- Berjalan paralel, tanpa mengirim data keluar (semua lokal).

## Instalasi

```sh
go install github.com/Emil-02/jaga/cmd/jaga@latest
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
jaga check . --ci --fail-on high   # untuk pipeline
```

Keluar dengan kode `1` bila ada temuan pada/melebihi `--fail-on` (default: `high`).

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
- uses: Emil-02/jaga@v0.1.0
  with:
    path: .
    format: sarif
    fail-on: high
```

## Arsitektur

```
cmd/jaga        entrypoint CLI (stdlib flag, tanpa dependensi eksternal)
internal/check  implementasi pemeriksaan + interface umum
internal/runner eksekusi paralel + timeout
internal/config parser .jaga.yml (subset yang didokumentasikan)
internal/report formatter text/json/sarif/badge
internal/score  perhitungan skor 0-100
```

Rencana lanjutan: integrasi penuh `leakboy` dan `gh-spruce`, rule packs dari komunitas, dan dukungan konfigurasi YAML penuh.

## Lisensi & atribusi

MIT. Dibuat dan dipelihara oleh [@Emil-02](https://github.com/Emil-02). Seluruh kepemilikan dan atribusi kode di repository ini mengacu pada akun GitHub tersebut (lihat [`.github/CODEOWNERS`](.github/CODEOWNERS)).
