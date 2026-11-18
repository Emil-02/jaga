# Arsitektur

`jaga-skill` bukan aplikasi yang dijalankan terus-menerus. Produknya adalah **file instruksi yang dipasang ke repo user**, plus installer kecil yang mengelola siklus hidupnya.

## Lapisan

```
src/cli.mjs         → argumen & perintah (init, check, doctor, uninstall, update)
src/detect.mjs      → deteksi agent yang terpasang
src/targets/*.mjs   → adapter: tahu ke mana file ditulis per agent
src/install.mjs     → render, tulis atomik, jaga idempotensi
src/manifest.mjs    → .jaga-skill.json (path, hash, versi)
src/binary.mjs      → resolver scanner (jaga → go → release → JS fallback)
src/hook.mjs        → git pre-commit opsional bermarker
src/config.mjs      → jaga.skill.yml
src/js-scanner.mjs  → fallback lokal tanpa Go
```

## Kontrak target

Tiap adapter mengekspor:

```js
export const target = {
  id,
  label,
  async detect(root) {},
  files(ctx) { return [{ rel, content, kind: 'skill' | 'section' }] },
}
```

- `skill` → file penuh, tidak pernah ditimpa bila berbeda (konflik → peringatan).
- `section` → blok `<!-- jaga:begin --> … <!-- jaga:end -->`, ditambahkan/diganti di dalam file yang sudah ada.

## Manifest

```json
{
  "version": "0.1.0",
  "targets": ["opencode", "claude"],
  "files": [
    { "path": ".opencode/skills/jaga/SKILL.md", "hash": "sha256:…", "kind": "skill" }
  ]
}
```

Uninstall membaca manifest dan menghapus **hanya** path yang tercatat. File milik user tidak pernah disentuh.

## Keamanan

- Tidak ada jaringan saat `init`/`check` kecuali resolver memilih mengunduh rilis.
- Unduhan (jika terjadi) diverifikasi SHA-256.
- Tidak ada telemetri.
- Secret/PII disensor sebelum ditampilkan.
