# Security Policy

`jaga-skill` hanya menulis file instruksi ke direktori project. Ia tidak mengumpulkan data, tidak mengirim telemetri, dan tidak menyimpan kredensial. Scanner yang dipakai (`jaga`) berjalan lokal dan menyensor secret/PII di keluaran.

Ketika memasang git hook, `jaga-skill` hanya menambahkan blok bermarker ke hook yang sudah ada — tidak menimpa isi user.

## Melaporkan kerentanan

Jangan buka issue publik. Laporkan lewat
[GitHub Security Advisories](https://github.com/Emil-02/jaga-skill/security/advisories/new).
Respons dalam 7 hari.

## Dukungan

Rilis `0.1.x` mendapat perbaikan keamanan.
