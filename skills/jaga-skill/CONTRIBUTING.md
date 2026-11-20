# Berkontribusi

Terima kasih sudah mau bantu mengembangkan `jaga-skill`. Saya senang menerima perubahan kecil maupun besar, selama masih sesuai arah proyeknya.

Repository ini saya kelola lewat akun [@Emil-02](https://github.com/Emil-02). Pemilik dan reviewer untuk semua perubahan ada di [`.github/CODEOWNERS`](../../.github/CODEOWNERS).

## Alur kerja

1. Buka issue dulu kalau perubahannya besar — misalnya dukungan agent baru atau skema konfigurasi baru.
2. Buat branch dari `main`, misalnya `target/nama-agent`, `fix/xxx`, atau `docs/xxx`.
3. Tambah test di `test/` untuk setiap perilaku baru.
4. Jalankan `npm test` dan `node --check src/cli.mjs` sebelum buat PR.
5. Di PR, ceritakan masalahnya, solusinya, dan cara mengujinya.

## Menambah target agent

Buat `src/targets/<id>.mjs` dengan ekspor seperti ini:

```js
export const target = {
  id: 'nama-agent',
  label: 'Nama Agent',
  async detect(root) {},
  files(ctx) { return [{ rel: '...', content: '...', kind: 'skill' | 'section' }] },
}
```

Lalu daftarkan di `src/detect.mjs`. Target baru perlu fixture dan test idempotensi.

## Isi `SKILL.md`

Saya sengaja menjaga instruksinya tetap pendek dan tegas. Detail teknisnya saya taruh di `references/`. Aturannya sederhana: jangan pernah minta agent membocorkan secret atau data pribadi, dan jangan minta dia mengirim file keluar.
