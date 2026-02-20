# Agent Service Prototype

Prototype layanan agent berbasis Go dengan HTTP API (Echo), database PostgreSQL, dan Ent ORM. Mendukung alur setup/installation dari bundle (manifest + migrations) dan health check.

## Persyaratan

- **Go** 1.25+
- **PostgreSQL** (untuk database)

## Struktur Proyek

```
agent-service-prototype/
├── cmd/server/          # Entry point aplikasi
├── internal/
│   ├── app/agent-service-prototype/   # Modul utama (controller, service, repository, dto, routes)
│   ├── bootstrap/       # Inisialisasi database
│   ├── config/          # Load konfigurasi dari environment
│   ├── router/          # Registrasi HTTP routes
│   └── server/          # HTTP server & graceful shutdown
├── pkg/
│   ├── logger/          # Zerolog logger
│   ├── setup/           # Bundle, manifest, migrations, status setup
│   └── utils/           # Utilitas umum
└── ent/schema/          # Ent schema (contoh: ExampleEntity)
```

## Konfigurasi (Environment)

Buat file `.env` di root proyek. Variabel **wajib**:

| Variabel     | Deskripsi                    |
|-------------|------------------------------|
| `APP_ENV`   | Environment (dev/staging/prod)|
| `APP_PORT`  | Port HTTP server             |
| `BUNDLE_URL`| URL bundle untuk setup DB    |
| `DB_HOST`   | Host PostgreSQL              |
| `DB_PORT`   | Port PostgreSQL              |
| `DB_USER`   | User database                |
| `DB_PASSWORD` | Password database          |
| `DB_NAME`   | Nama database                |
| `DB_SSL_MODE` | SSL mode (disable/require dll) |

Variabel **opsional** (punya default):

| Variabel | Default | Deskripsi |
|----------|---------|-----------|
| `WORK_DIR` | `./.work` | Direktori kerja (download bundle, dll) |
| `ADVISORY_LOCK_KEY` | `987654321` | Kunci advisory lock |
| `FORCE` | `false` | Force installation |
| `SKIP_SMOKE` | `false` | Skip smoke check |

Contoh `.env`:

```env
APP_ENV=dev
APP_PORT=8080
BUNDLE_URL=https://example.com/db-bundle.zip
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=agent_db
DB_SSL_MODE=disable
```

## Menjalankan Aplikasi

1. **Install dependensi**

   ```bash
   go mod download
   ```

2. **Set environment** (atau gunakan `.env`)

   Pastikan semua variabel wajib sudah di-set.

3. **Jalankan server**

   ```bash
   go run ./cmd/server
   ```

   Server HTTP akan listen di `http://localhost:<APP_PORT>`.

## API HTTP

- **GET /health** — Health check (tanpa auth).  
  Response: `{"status":"ok"}`

- **GET /setup/status** — Status proses setup/installation.  
  Response: status (idle/running/success/failed), step, error, started_at, finished_at.

- **POST /setup/installation** — Menjalankan installation dari bundle (download, extract, manifest, baseline, migrations, smoke).  
  - 200: success/failed (lihat body).  
  - 409: installation sudah berjalan (conflict).

## Generate Kode

### Ent (ORM)

Setelah schema di `ent/schema/` siap, generate client Ent:

```bash
go generate ./ent
```

Atau:

```bash
go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema
```

Setelah generate, uncomment kode yang memakai Ent di `internal/bootstrap`, `internal/router`, dan `internal/app/agent-service-prototype/routes` sesuai kebutuhan.

## Tech Stack

- **HTTP:** [Echo v4](https://echo.labstack.com/)
- **Database:** PostgreSQL ([lib/pq](https://github.com/lib/pq)), [Ent](https://entgo.io/)
- **Logging:** [zerolog](https://github.com/rs/zerolog)
- **Config:** [godotenv](https://github.com/joho/godotenv)

## Lisensi

Private / internal — sesuaikan dengan kebijakan tim.
