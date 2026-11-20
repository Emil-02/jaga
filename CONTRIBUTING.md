# Berkontribusi ke jaga

Terima kasih sudah tertarik. Sebelum mulai, baca dulu singkatnya di sini supaya prosesnya mulus.

## Aturan atribusi

Repository ini dibuat dan dipelihara oleh akun GitHub [@Emil-02](https://github.com/Emil-02). Seluruh atribusi, kepemilikan, dan kredit penulisan di repository ini tercatat pada akun tersebut — tidak ada pencantuman co-author, co-contributor, atau nama/akun lain dalam dokumentasi, metadata, maupun riwayat atribusi resmi proposal. Lihat [`.github/CODEOWNERS`](.github/CODEOWNERS). Jika kamu mengirim perubahan, diskusikan dulu lewat issue/PR agar atribusinya tetap sesuai alur ini.

## Alur kerja

1. Buka issue dulu untuk perubahan besar (check baru, format report baru, perubahan skema config).
2. Fork/repo ini, buat branch dari `main` dengan nama `check/xxx`, `fix/xxx`, atau `docs/xxx`.
3. Tulis test untuk setiap perilaku baru.
4. Jalankan pemeriksaan sebelum buat PR:

```sh
gofmt -w cmd internal
go vet ./cmd/... ./internal/...
go test ./cmd/... ./internal/...
go build ./cmd/...
```

5. Buat PR dengan deskripsi: masalah, solusi, dan cara menguji.

## Menambah check baru

Implementasi `check.Check` di `internal/check`:

```go
type Check interface {
    Name() string
    Run(ctx context.Context, root string, cfg config.Config) Result
}
```

- Jangan pernah mencatat nilai secret/PII mentah — selalu pakai `mask()`.
- Hormati `cfg.Exclude`.
- Gunakan `Severity` yang sudah ada: `low`, `medium`, `high`, `critical`.
- Daftarkan check di `runner.NewDefault()` dan tambahkan test.

## Format pesan commit

`type: deskripsi singkat` — contoh: `check: deteksi token OpenAI`.

## Kode etik

Bersikap baik, fokus pada teknis, dan jangan memposting data pribadi atau secret di issue/PR. Gunakan tooling `sensor` bila perlu menyensor contoh data.
