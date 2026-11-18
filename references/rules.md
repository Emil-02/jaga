# Aturan pemeriksaan

Ringkasan aturan yang dipakai `jaga`. Detail implementasi ada di repository `jaga`.

## secret (critical)

- AWS access key `AKIA…`
- GitHub token `ghp_`, `gho_`, `ghu_`, `ghs_`, `ghr_`
- OpenAI key `sk-…`, `sk-proj-…`
- Slack `xox…`, Stripe live `sk_live_…`
- Private key PEM
- Credential generik: `api_key`, `secret`, `token`, `password` dengan nilai panjang

## pii (high)

- NIK / KK: 16 digit
- NPWP: `00.000.000.0-000.000`
- Nomor HP Indonesia: `08…`
- Nomor rekening setelah kata `rekening` / `no. rek`

## hygiene (low)

- `LICENSE`, `README`, workflow CI, `.gitignore`

## endpoint (rencana)

- Supabase / Firebase / Stripe / OpenAI key yang terekspos di bundle
- Diaktifkan dengan mode `--deep` pada `jaga`

## Penanganan

Semua nilai yang cocok di-mask sebelum ditampilkan. File `_test.go`, `testdata`, `node_modules`, `dist`, `vendor`, dan `.git` dilewati.
