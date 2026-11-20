# Menambahkan Rule Baru di Jaga 🌿

Panduan ini dibuat bagi siapa saja yang ingin berkontribusi menambahkan aturan (*rules*) baru ke `jaga`, baik untuk deteksi API key baru, platform AI, fintech, maupun data sensitif lainnya.

---

## 1. Struktur Modul Pemeriksaan

Pemeriksaan `jaga` berada di dalam direktori `internal/check/`:

- `secret.go` — Deteksi token, API key, kredensial cloud, AI services, dan secret kebocoran frontend.
- `pii.go` — Deteksi Data Pribadi (Personal Identifiable Information) seperti NIK, NPWP, nomor rekening, dan nomor telepon.
- `hygiene.go` — Pemeriksaan kelengkapan repo standar (`LICENSE`, `README`, `.gitignore`, workflow CI).

Setiap modul mengimplementasikan antarmuka `Check`:
```go
type Check interface {
    Name() string
    Run(ctx context.Context, root string, cfg config.Config) Result
}
```

---

## 2. Cara Menambahkan Rule Secret Baru

Buka file `internal/check/secret.go`. Tambahkan entri baru ke dalam slice `secretRules`:

```go
type secretRule struct {
    name     string
    re       *regexp.Regexp
    validate func(match string, lineText string) bool // Opsional: logika validasi khusus
}
```

### Contoh 1: Menambahkan API Key Sederhana dengan Regex
```go
{
    name: "Provider XYZ API key",
    re:   regexp.MustCompile(`\bxyz_[A-Za-z0-9]{32,}\b`),
},
```

### Contoh 2: Menambahkan Rule dengan Validasi Khusus
Jika rule kamu memerlukan validasi khusus (misal: verifikasi payload JWT, checksum, atau pengecualian token publik):

```go
{
    name: "Supabase service_role key",
    re:   regexp.MustCompile(`\b(sb_secret_[A-Za-z0-9_-]{20,}|eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,})\b`),
    validate: func(m string, _ string) bool {
        if strings.HasPrefix(m, "sb_secret_") {
            return !isPlaceholder(m)
        }
        return isSupabaseServiceRole(m)
    },
},
```

> **Catatan Masking & Placeholder:**
> Semua rule secara otomatis disaring oleh helper `isPlaceholder(m)` untuk menghindari *false positive* dari dokumentasi contoh seperti `your_api_key_here` atau `dummy_key`. Semua temuan yang lolos akan dimasking otomatis oleh `Mask(m)` saat ditampilkan ke laporan.

---

## 3. Menambahkan Unit Test

Setiap rule baru **wajib** dilengkapi unit test di `internal/check/check_test.go`.

Tambahkan *test case* ke dalam fungsi `TestSecretVibeCodingAndAI` atau buat fungsi test baru:

```go
func TestNewProviderRule(t *testing.T) {
    dir := t.TempDir()
    content := "XYZ_KEY=xyz_abcdef1234567890abcdef1234567890\n"
    if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o600); err != nil {
        t.Fatal(err)
    }

    res := Secret{}.Run(context.Background(), dir, defaultCfg())
    if len(res.Findings) == 0 {
        t.Fatal("secret XYZ tidak terdeteksi")
    }
}
```

---

## 4. Menjalankan Tes & Verifikasi

Sebelum mengirim Pull Request, pastikan seluruh test lokal lulus:

```bash
# Jalankan test khusus check
go test -v ./internal/check

# Jalankan seluruh test suite
go test ./cmd/... ./internal/...
```

---

## 5. Checklist Pull Request

- [ ] Rule memiliki regex yang spesifik dengan pembatas kata (`\b`) agar tidak memicu *false positive*.
- [ ] Rule tidak membocorkan plaintext di error/message (gunakan masking default).
- [ ] Dilengkapi unit test di `internal/check/check_test.go`.
- [ ] Semua unit test lulus dengan `go test ./...`.
