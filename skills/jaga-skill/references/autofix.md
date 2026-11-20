# Batas auto-fix

## Boleh tanpa tanya

- Menambah `LICENSE`, `README.md`, `.gitignore`, workflow CI minimal
- Menambah entri ke `exclude` untuk direktori build yang jelas (`dist`, `build`, `vendor`)
- Mengubah format laporan agar lebih mudah dibaca agent

## Harus tanya dulu

- Menghapus atau mengganti nilai secret / API key → minta rotasi dari user
- Mengubah data pribadi (NIK, NPWP, rekening, nomor HP)
- Mematikan check, menaikkan `--fail-on`, atau menghapus git hook
- Menimpa file yang sudah ada di project

## Dilarang

- Print nilai asli temuan
- Commit/push atas nama user tanpa konfirmasi
- Mengirim isi file ke layanan pihak ketiga
