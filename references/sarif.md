# Membaca laporan SARIF

`jaga-skill check --format sarif` menghasilkan SARIF 2.1.0.

## Struktur penting

```json
{
  "version": "2.1.0",
  "runs": [{
    "tool": { "driver": { "name": "jaga" } },
    "results": [{
      "ruleId": "secret",
      "level": "error",
      "message": { "text": "AWS access key: AKIA…CDEF" },
      "locations": [{
        "physicalLocation": {
          "artifactLocation": { "uri": "src/config.js" },
          "region": { "startLine": 12 }
        }
      }]
    }]
  }]
}
```

## Prioritas

| `level` | Arti | Tindakan |
|---|---|---|
| `error` | `critical` / `high` | Blokir commit/deploy |
| `warning` | `medium` | Perbaiki sebelum rilis |
| `note` | `low` | Boleh menyusul |

## Catatan

- `message.text` sudah disensor — aman ditampilkan.
- `ruleId = hygiene` biasanya aman di-auto-fix.
- `ruleId = secret` tidak boleh diperbaiki dengan sekadar menghapus baris; kredensial harus dirotasi.
