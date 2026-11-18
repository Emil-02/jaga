# Berkontribusi

Repository ini dibuat dan dipelihara oleh akun GitHub [@Emil-02](https://github.com/Emil-02). Seluruh atribusi dan kepemilikan tercatat pada akun tersebut (lihat [`.github/CODEOWNERS`](.github/CODEOWNERS)); tidak ada pencantuman co-author, co-contributor, atau akun lain dalam metadata maupun dokumentasi resmi.

## Alur

1. Issue dulu untuk perubahan besar (target agent baru, format output, skema config).
2. Branch dari `main`: `target/nama-agent`, `fix/xxx`, `docs/xxx`.
3. Tambah test di `test/` untuk setiap perilaku baru.
4. Jalankan `npm test` dan `node --check src/cli.mjs` sebelum PR.
5. PR berisi: masalah, solusi, cara menguji.

## Menambah target agent

Buat `src/targets/<id>.mjs` dengan ekspor:

```js
export const target = {
  id: 'nama-agent',
  label: 'Nama Agent',
  detect: async (root) => boolean,
  files: (ctx) => [{ rel: '...', content: '...', kind: 'skill' | 'section' }],
}
```

Daftarkan di `src/detect.mjs`. Target baru wajib punya fixture dan test idempotensi.

## Aturan konten `SKILL.md`

- Instruksi harus singkat dan deterministik.
- Detail pindah ke `references/`.
- Dilarang meminta agent membocorkan secret/PII mentah atau mengirim data keluar.
