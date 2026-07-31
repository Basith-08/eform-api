# E-Form Employee Management API

Backend REST API untuk sistem manajemen data karyawan. Aplikasi dibangun dengan Go, Gin, GORM, dan PostgreSQL.

## Fitur

- Registrasi, login, refresh token, logout, dan reset password
- Profil karyawan dan perhitungan kelengkapan profil
- Dashboard user dan admin
- CRUD karyawan untuk admin
- Upload dokumen ke local storage
- JWT authentication, role-based authorization, CORS, secure headers, dan rate limiting
- Migrasi serta seed SQL otomatis saat aplikasi mulai
- Dokumentasi API melalui Swagger

## Teknologi

- Go 1.25+
- Gin
- GORM dengan PostgreSQL
- JWT (`github.com/golang-jwt/jwt/v5`)
- Docker dan Docker Compose

## Struktur Direktori

```text
cmd/api/          Entry point aplikasi
config/           Pembacaan environment variable
internal/
  domain/         Model domain
  handler/        HTTP handler
  middleware/     Authentication dan middleware HTTP
  repository/     Akses database
  routes/         Registrasi route
  service/        Business logic
  validator/      Validasi request
migrations/       SQL schema migration
seeds/            Data awal
pkg/              Package reusable (database, JWT, storage, response)
tests/            Automated tests
```

## Menjalankan Secara Lokal

Prasyarat:

- Go 1.25 atau lebih baru
- PostgreSQL yang dapat diakses aplikasi

```bash
cp .env.example .env
```

Untuk menjalankan binary dari host, ubah `DB_HOST` di `.env` menjadi `localhost` dan sesuaikan kredensial PostgreSQL. Setelah database tersedia:

```bash
go run ./cmd/api
```

Aplikasi berjalan di `http://localhost:8080` secara default.

Pada startup, aplikasi menjalankan file `.sql` dalam `migrations/` dan `seeds/` secara berurutan. Seed dapat dinonaktifkan dengan:

```env
SEED_ON_BOOT=false
```

## Environment Variable

Salin `.env.example` sebagai titik awal. Variable utama yang digunakan aplikasi:

| Variable | Default | Keterangan |
|---|---|---|
| `APP_ENV` | `development` | Environment aplikasi |
| `APP_NAME` | `E-Form Employee Management System` | Nama service |
| `PORT` | `8080` | Port HTTP |
| `FRONTEND_URL` | `http://localhost:5173` | Origin yang diizinkan CORS |
| `UPLOAD_PATH` | `./uploads` | Direktori upload dokumen |
| `MAX_UPLOAD_BYTES` | `1048576` | Batas ukuran upload dalam byte |
| `RATE_LIMIT_RPS` | `10` | Request per detik |
| `RATE_LIMIT_BURST` | `20` | Burst rate limit |
| `SEED_ON_BOOT` | `true` | Jalankan seed saat startup |
| `DB_HOST` | `localhost` | Host PostgreSQL |
| `DB_PORT` | `5432` | Port PostgreSQL |
| `DB_NAME` | `eform` | Nama database aplikasi |
| `DB_USER` | `postgres` | User database aplikasi |
| `DB_PASSWORD` | `postgres` | Password database aplikasi |
| `DB_SSLMODE` | `disable` | SSL mode PostgreSQL |
| `DB_TIMEZONE` | `UTC` | Time zone koneksi database |
| `JWT_SECRET` | `change-this-secret` | Secret signing JWT; wajib diganti di deployment |
| `ACCESS_TOKEN_TTL` | `15m` | Masa berlaku access token |
| `REFRESH_TOKEN_TTL` | `168h` | Masa berlaku refresh token |

Jangan commit `.env` yang berisi secret.

## API

Health check:

```text
GET /health
```

Swagger UI tersedia di:

```text
GET /swagger/index.html
```

Base path API adalah `/api/v1`.

| Kelompok | Endpoint |
|---|---|
| Auth publik | `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/forgot-password`, `/auth/reset-password` |
| Profil | `GET /profile/me`, `PUT /profile/me`, `POST /auth/change-password` |
| Dashboard | `GET /dashboard/user`, `GET /dashboard/admin` (admin) |
| Karyawan | `GET/POST /employees`, `GET/PUT/DELETE /employees/:id` (admin) |
| Status karyawan | `PATCH /employees/:id/activate`, `PATCH /employees/:id/deactivate` (admin) |
| Reset password admin | `POST /employees/:id/reset-password` (admin) |

Endpoint protected membutuhkan header:

```http
Authorization: Bearer <access-token>
```

## Testing dan Build

```bash
go test ./...
go build ./cmd/api
```

CI menjalankan kedua command tersebut sebelum image Docker dibuat.

## Docker

Build image:

```bash
docker build -t eform-api .
```

Image menjalankan binary `/app/eform-api`, mengekspos port `8080`, dan membawa direktori `migrations/` serta `seeds/`.

`compose.yaml` ditujukan untuk deployment image dari GHCR, bukan development lokal. Variable deployment yang diperlukan antara lain `GHCR_OWNER`, `IMAGE_TAG`, dan `PLATFORM_DOMAIN`. Service PostgreSQL Compose juga membutuhkan variable `POSTGRES_USER`, `POSTGRES_PASSWORD`, dan `POSTGRES_DB` selain variable `DB_*` yang digunakan aplikasi.

Compose memakai external network `edge` serta bind mount production di `/srv/apps/eform-api/volumes/`; pastikan keduanya tersedia sebelum menjalankannya.

## Deployment

Workflow `.github/workflows/deploy.yml` berjalan ketika ada push ke branch `main` atau dijalankan manual untuk redeploy SHA tertentu. Alurnya:

1. Menjalankan test dan build Go.
2. Build dan push image ke GHCR dengan tag commit SHA.
3. Mengubah `IMAGE_TAG` di server production.
4. Menjalankan `docker compose pull` dan `docker compose up -d` melalui SSH.

