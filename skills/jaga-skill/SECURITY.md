# Security Policy

`jaga-skill` cuma menulis file instruksi ke direktori project. Tidak ada telemetri, tidak ada data yang dikirim keluar, dan tidak ada kredensial yang disimpan. Scanner yang dipakai berjalan lokal dan menyensor secret/PII di keluaran.

Kalau Anda memakai git hook, `jaga-skill` hanya menambahkan blok bermarker ke hook yang sudah ada — isi hook punya Anda tetap utuh.

## Melaporkan kerentanan

Sebaiknya jangan buka issue publik. Kirim laporannya lewat [GitHub Security Advisories](https://github.com/Emil-02/jaga-skill/security/advisories/new) supaya detailnya tetap privat. Saya biasanya merespons dalam 7 hari.

## Dukungan

Rilis `0.1.x` masih saya dukung dengan perbaikan keamanan.
