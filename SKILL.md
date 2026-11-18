---
name: jaga
description: >
  Security and privacy gate for AI coding agents. Use this skill before
  committing, pushing, deploying, or sharing code. Scans for leaked credentials,
  Indonesian personal data, exposed endpoints and repo hygiene, returns SARIF,
  and blocks when severity exceeds the configured threshold.
license: MIT
---

# jaga — security gate

Versi {{VERSION}}. Dibuat dan dipelihara oleh [@Emil-02](https://github.com/Emil-02).

## Kapan dipakai

Gunakan skill ini **setiap kali** kamu akan:

- `git commit` atau `git push`
- deploy / build produksi
- membuat atau mengubah file `.env`, konfigurasi API, atau kredensial
- membagikan screenshot, log, atau cuplikan kode ke orang lain

Jangan hanya mengandalkan insting: jalankan pemeriksaan.

## Aturan keras

1. **Jangan pernah mencetak nilai secret atau data pribadi mentah** ke terminal, komentar kode, atau commit message. `jaga` sudah menyensornya — pertahankan itu.
2. **Jangan pernah auto-push atau auto-deploy** setelah menemukan secret. Rotasi kredensial dulu, lalu minta konfirmasi user.
3. **Jangan mematikan gate** (`--fail-on`, exclude, atau hapus hook) tanpa izin eksplisit user.
4. **Jangan mengirim file ke layanan eksternal** untuk "diperiksa". Semua pemeriksaan berjalan lokal.
5. Jika binary `jaga` tidak tersedia, pakai fallback bawaan: `jaga-skill check` otomatis beralih ke scanner JS.

## Alur kerja

1. Jalankan `jaga-skill check` di root repo.
2. Baca keluaran SARIF / teks. Urutannya: `critical` → `high` → `medium` → `low`.
3. Perbaiki temuan:
   - **hygiene** (LICENSE, README, CI, .gitignore) → boleh diperbaiki sendiri.
   - **secret** → minta user merotasi kredensial; jangan sekadar menghapus baris.
   - **PII** → tanya apakah datanya memang perlu disimpan; sensor bila perlu.
4. Jalankan `jaga-skill check` lagi sampai lolos.
5. Baru lanjut ke commit/deploy.

## Perintah

```sh
jaga-skill check                       # teks, ringkas
jaga-skill check --format sarif        # untuk dibaca agent
jaga-skill check --fail-on high        # ambang kegagalan
jaga-skill doctor                      # cek instalasi & binary
```

Kode keluar `0` = lolos, `1` = ada temuan di atas ambang, `2` = error konfigurasi.

## Membaca SARIF

Pada `runs[].results[]`:

- `ruleId` → jenis check (`secret`, `pii`, `hygiene`, `endpoint`)
- `level` → `error` = harus diperbaiki, `warning` → sebaiknya, `note` → opsional
- `locations[0].physicalLocation` → file dan baris yang perlu disentuh
- `message.text` → sudah disensor; aman ditampilkan ke user

## Auto-fix policy

| Temuan | Boleh agent perbaiki? |
|---|---|
| LICENSE / README / CI / .gitignore hilang | Ya |
| exclude path salah | Ya, setelah konfirmasi singkat |
| secret / API key | Tidak — rotasi dulu oleh user |
| NIK / NPWP / rekening | Tidak — tanya dulu |

## Referensi

Baca hanya jika diperlukan: `references/rules.md`, `references/sarif.md`, `references/autofix.md`.
