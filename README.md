# API Students

REST API untuk mengelola data mahasiswa menggunakan Go, Fiber v2, dan PostgreSQL.

Project ini dibuat untuk memenuhi tugas mandiri Pertemuan 3 — Database & Repository Pattern.

Pada Pertemuan 3, penyimpanan data mahasiswa yang sebelumnya dilakukan di dalam memory dipindahkan ke PostgreSQL. Perilaku HTTP API tetap dipertahankan, sedangkan perubahan utama terdapat pada tempat penyimpanan data dan penggunaan Repository Pattern.

---

## Teknologi

- Go
- Fiber v2
- PostgreSQL
- REST API
- JSON
- HTTP
- PostgreSQL Connection Pool

---

## Struktur Project

```text
api-students/
├── app/
│   └── repository/
│       └── student_repository.go
├── database/
│   └── postgres.go
├── migrations/
│   └── 001_create_students.sql
├── model/
│   └── student.go
├── main.go
├── handler.go
├── helper.go
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
└── README.md
```

### Penjelasan File

| File | Keterangan |
|---|---|
| `main.go` | Menjalankan server, middleware, routing API, dan koneksi database |
| `handler.go` | Berisi handler untuk endpoint API |
| `helper.go` | Berisi helper untuk membuat response berhasil dan response error |
| `model/student.go` | Berisi struct Student dan struct request untuk POST, PUT, dan PATCH |
| `database/postgres.go` | Membuat connection pool PostgreSQL dan melakukan Ping |
| `app/repository/student_repository.go` | Berisi interface dan implementasi StudentRepository untuk PostgreSQL |
| `migrations/001_create_students.sql` | Berisi schema tabel students dan index |
| `.env.example` | Template variabel environment yang dibutuhkan aplikasi |
| `.gitignore` | Mengatur file yang tidak boleh di-commit, termasuk `.env` |
| `README.md` | Dokumentasi project dan API |

---

# Base URL

```text
http://localhost:3000/api/v1
```

Health check:

```text
http://localhost:3000/health
```

---

# Database PostgreSQL

Pada project ini, data student disimpan menggunakan PostgreSQL.

Sebelumnya data student disimpan di dalam memory aplikasi. Pada Pertemuan 3, penyimpanan tersebut dipindahkan ke database PostgreSQL.

Perubahan ini tidak mengubah kontrak HTTP API. Endpoint, method HTTP, request, response, dan status HTTP tetap digunakan seperti sebelumnya.

---

# Skema Database

Database menggunakan tabel:

```text
students
```

Struktur tabel:

| Kolom | Tipe | Constraint / Keterangan |
|---|---|---|
| `id` | `SERIAL` | Primary key |
| `nim` | `VARCHAR(20)` | Wajib diisi dan unik |
| `name` | `VARCHAR(100)` | Wajib diisi |
| `grade` | `NUMERIC(5,2)` | Nilai mahasiswa |
| `is_active` | `BOOLEAN` | Status aktif, default `TRUE` |
| `created_at` | `TIMESTAMP` | Waktu data dibuat |

Migration database terdapat pada:

```text
migrations/001_create_students.sql
```

Isi migration:

```sql
CREATE TABLE students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_students_name ON students(name);
```

---

# Keunikan NIM

Kolom `nim` menggunakan constraint:

```sql
UNIQUE
```

NIM wajib unik karena setiap mahasiswa harus memiliki identitas NIM yang berbeda.

Keunikan NIM lebih baik dijaga oleh database daripada hanya oleh kode Go karena database merupakan lapisan yang secara langsung menjaga integritas data.

Jika hanya dilakukan melalui kode Go, masih terdapat kemungkinan dua request datang secara bersamaan dan keduanya melewati pengecekan sebelum salah satunya melakukan insert.

Dengan constraint `UNIQUE`, PostgreSQL akan tetap menolak data dengan NIM yang sama meskipun data dimasukkan melalui proses atau request yang berbeda.

Jika terjadi NIM ganda, aplikasi menerjemahkan error database menjadi:

```text
409 Conflict
```

---

# Index

Selain primary key, tabel `students` memiliki index:

```sql
CREATE INDEX idx_students_name ON students(name);
```

Index tersebut dibuat pada kolom `name` karena API memiliki fitur pencarian mahasiswa berdasarkan nama.

Index dapat membantu database dalam mencari data berdasarkan kolom tertentu, terutama ketika jumlah data semakin besar.

---

# Environment Configuration

Kredensial PostgreSQL disimpan menggunakan environment variable melalui file:

```text
.env
```

File `.env` tidak boleh di-commit ke GitHub karena dapat berisi username, password, dan informasi koneksi database.

Sebagai gantinya, project menyediakan:

```text
.env.example
```

---

# Variabel Environment

Variabel yang diperlukan:

```env
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_SSLMODE=
```

Contoh konfigurasi lokal:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=api_students
DB_SSLMODE=disable
```

Sesuaikan nilai tersebut dengan konfigurasi PostgreSQL pada komputer masing-masing.

---

# Persiapan Database dari Nol

## 1. Pastikan PostgreSQL Terinstall

Periksa versi PostgreSQL:

```powershell
psql --version
```

Jika PostgreSQL sudah terinstall, terminal akan menampilkan versi PostgreSQL.

---

## 2. Buat Database

Masuk ke PostgreSQL menggunakan user `postgres`:

```powershell
psql -U postgres
```

Buat database:

```sql
CREATE DATABASE api_students;
```

Setelah selesai, keluar dari PostgreSQL:

```sql
\q
```

---

## 3. Buat File `.env`

Buat file `.env` pada root project:

```text
api-students/
├── .env
├── main.go
├── handler.go
├── helper.go
└── ...
```

Isi `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=api_students
DB_SSLMODE=disable
```

Password disesuaikan dengan password PostgreSQL masing-masing.

---

## 4. Jalankan Migration

Migration tersedia pada:

```text
migrations/001_create_students.sql
```

Jalankan migration:

```powershell
psql -U postgres -d api_students -f migrations/001_create_students.sql
```

Jika berhasil, tabel `students` dan index `idx_students_name` akan dibuat.

---

## 5. Memeriksa Tabel

Masuk ke database:

```powershell
psql -U postgres -d api_students
```

Periksa tabel:

```sql
\dt
```

Periksa struktur tabel:

```sql
\d students
```

Periksa index:

```sql
\di
```

---

# Connection Pool PostgreSQL

Koneksi PostgreSQL dibuat pada:

```text
database/postgres.go
```

File tersebut bertanggung jawab untuk:

1. Membaca konfigurasi database dari environment variable.
2. Membuat connection pool.
3. Melakukan `Ping` ke PostgreSQL.
4. Mengembalikan error jika database tidak dapat dihubungi.

Connection pool digunakan agar aplikasi dapat mengelola koneksi database secara efisien ketika menerima banyak request.

---

# Health Check

Endpoint:

```http
GET /health
```

Digunakan untuk memeriksa kondisi aplikasi dan koneksi PostgreSQL.

Health check tidak hanya memeriksa apakah server Go berjalan, tetapi juga melakukan pengecekan koneksi ke database.

Jika PostgreSQL tersedia:

```text
200 OK
```

Jika PostgreSQL tidak tersedia:

```text
503 Service Unavailable
```

Status `503 Service Unavailable` dipilih karena aplikasi masih berjalan, tetapi service database yang menjadi dependency utama sedang tidak tersedia.

---

# Repository Pattern

Data student tidak lagi disimpan langsung di handler atau memory.

Project menggunakan Repository Pattern untuk memisahkan proses HTTP dengan proses akses database.

Interface `StudentRepository` memiliki minimal lima method:

```text
FindAll
FindByID
Create
Update
Delete
```

Implementasi repository PostgreSQL terdapat pada:

```text
app/repository/student_repository.go
```

Repository tidak memiliki ketergantungan terhadap Fiber.

Dengan demikian, repository hanya bertanggung jawab terhadap operasi data dan database, sedangkan handler bertanggung jawab terhadap request dan response HTTP.

---

# Sentinel Error

Repository menyediakan dua sentinel error:

```text
ErrNotFound
ErrDuplicate
```

## ErrNotFound

Digunakan ketika data student yang diminta tidak ditemukan di PostgreSQL.

Error tersebut diterjemahkan menjadi:

```text
404 Not Found
```

Digunakan pada:

- GET student berdasarkan ID
- PUT student
- PATCH student
- DELETE student

---

## ErrDuplicate

Digunakan ketika terjadi konflik data, terutama ketika NIM yang digunakan sudah terdapat di database.

Error tersebut diterjemahkan menjadi:

```text
409 Conflict
```

Digunakan pada:

- POST student
- PUT student
- PATCH student

---

# Parameterized Query

Semua nilai yang berasal dari client dikirim ke PostgreSQL menggunakan parameter query.

Contoh:

```sql
SELECT id, nim, name, grade, is_active, created_at
FROM students
WHERE name ILIKE $1;
```

Nilai dari client tidak digabungkan langsung ke dalam string SQL.

Pendekatan ini digunakan untuk mencegah SQL Injection dan menjaga pemisahan antara query SQL dengan data dari client.

---

# Query di Sisi Database

Proses yang sebelumnya dilakukan di dalam Go pada Pertemuan 2 sekarang dipindahkan ke PostgreSQL.

Fitur yang diproses di database:

| Fitur | Implementasi PostgreSQL |
|---|---|
| Pencarian nama | `ILIKE` |
| Filtering | `WHERE` dengan parameter |
| Sorting | `ORDER BY` |
| Pagination | `LIMIT` dan `OFFSET` |
| Total data | `SELECT COUNT(*)` |

---

# Search

Pencarian nama menggunakan:

```sql
ILIKE
```

Contoh:

```http
GET /api/v1/students/?search=sAnIa
```

Pencarian tidak membedakan huruf besar dan kecil.

Contoh query:

```sql
WHERE name ILIKE $1
```

---

# Filtering

Filter menggunakan klausa:

```sql
WHERE
```

dengan parameter query.

## Filter Status Aktif

```http
GET /api/v1/students/?is_active=true
```

## Filter Grade Minimum

```http
GET /api/v1/students/?grade_min=80
```

## Filter Grade Maksimum

```http
GET /api/v1/students/?grade_max=95
```

## Filter Rentang Grade

```http
GET /api/v1/students/?grade_min=80&grade_max=95
```

---

# Sorting

Sorting menggunakan:

```sql
ORDER BY
```

Contoh:

```http
GET /api/v1/students/?sort=grade&order=desc
```

Data akan diurutkan berdasarkan grade dari terbesar ke terkecil.

Kolom yang dapat digunakan untuk sorting berasal dari daftar putih kolom yang telah ditentukan oleh aplikasi.

Input client tidak langsung dimasukkan sebagai nama kolom SQL.

Hal tersebut dilakukan untuk menjaga keamanan query dan mencegah manipulasi SQL melalui parameter sorting.

---

# Pagination

Pagination menggunakan:

```sql
LIMIT
OFFSET
```

Contoh:

```http
GET /api/v1/students/?page=1&limit=2
```

Nilai default:

```text
page = 1
limit = 10
```

Batas maksimum:

```text
limit = 50
```

Batas maksimum digunakan agar client tidak meminta jumlah data yang terlalu besar dalam satu response.

Perhitungan offset:

```text
offset = (page - 1) * limit
```

Contoh:

```text
page = 1
limit = 10

offset = (1 - 1) * 10
offset = 0
```

---

# Meta Pagination

Response daftar student memiliki metadata:

```json
{
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 3,
    "total_pages": 1
  }
}
```

| Field | Keterangan |
|---|---|
| `page` | Halaman yang sedang ditampilkan |
| `limit` | Jumlah data per halaman |
| `total` | Jumlah data yang sesuai dengan filter |
| `total_pages` | Jumlah seluruh halaman |

Nilai `total` diperoleh menggunakan:

```sql
SELECT COUNT(*)
```

dengan kondisi penyaringan yang sama dengan query pengambilan data.

---

# Struktur Data Student

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | int | ID unik student |
| `nim` | string | NIM unik mahasiswa |
| `name` | string | Nama mahasiswa |
| `grade` | numeric | Nilai mahasiswa |
| `is_active` | bool | Status aktif mahasiswa |
| `created_at` | timestamp | Waktu data dibuat |

Contoh:

```json
{
  "id": 1,
  "nim": "434241101",
  "name": "Sania",
  "grade": 91.5,
  "is_active": true,
  "created_at": "2026-08-30T10:00:00Z"
}
```

---

# API Contract

| Metode | Endpoint | Parameter | Contoh Body Permintaan | Status yang Mungkin Dikembalikan |
|---|---|---|---|---|
| GET | `/api/v1/students/` | `page`, `limit`, `search`, `sort`, `order`, `is_active`, `grade_min`, `grade_max` | Tidak ada | `200`, `400`, `500` |
| GET | `/api/v1/students/:id` | `id` | Tidak ada | `200`, `400`, `404`, `500` |
| POST | `/api/v1/students/` | Tidak ada | JSON student | `201`, `409`, `415`, `422`, `500` |
| PUT | `/api/v1/students/:id` | `id` | JSON student | `200`, `400`, `404`, `409`, `415`, `422`, `500` |
| PATCH | `/api/v1/students/:id` | `id` | JSON sebagian data | `200`, `400`, `404`, `409`, `415`, `422`, `500` |
| DELETE | `/api/v1/students/:id` | `id` | Tidak ada | `204`, `400`, `404`, `500` |

---

# Detail Endpoint

## 1. GET Daftar Student

### Endpoint

```http
GET /api/v1/students/
```

Digunakan untuk mengambil daftar student dari PostgreSQL.

### Query Parameter

| Parameter | Keterangan | Default | Contoh |
|---|---|---|---|
| `page` | Nomor halaman | `1` | `page=1` |
| `limit` | Jumlah data per halaman | `10` | `limit=2` |
| `search` | Pencarian berdasarkan nama | - | `search=sania` |
| `sort` | Field untuk sorting | - | `sort=grade` |
| `order` | Urutan `asc` atau `desc` | - | `order=desc` |
| `is_active` | Filter status aktif | - | `is_active=true` |
| `grade_min` | Nilai minimum grade | - | `grade_min=80` |
| `grade_max` | Nilai maksimum grade | - | `grade_max=95` |

### Contoh Request

```http
GET /api/v1/students/?page=1&limit=2
```

### Contoh Response

```json
{
  "success": true,
  "message": "Data student berhasil diambil",
  "data": [
    {
      "id": 1,
      "nim": "434241101",
      "name": "Sania",
      "grade": 91.5,
      "is_active": true
    }
  ],
  "meta": {
    "page": 1,
    "limit": 2,
    "total": 3,
    "total_pages": 2
  }
}
```

### Status

- `200 OK` — data berhasil diambil.
- `400 Bad Request` — parameter tidak valid.
- `500 Internal Server Error` — terjadi kesalahan server atau database.

---

## 2. GET Student Berdasarkan ID

### Endpoint

```http
GET /api/v1/students/:id
```

### Parameter

| Parameter | Keterangan |
|---|---|
| `id` | ID student berupa angka |

### Contoh Request

```http
GET /api/v1/students/1
```

### Status

- `200 OK` — student ditemukan.
- `400 Bad Request` — ID bukan angka.
- `404 Not Found` — student tidak ditemukan.
- `500 Internal Server Error` — terjadi kesalahan server atau database.

---

## 3. POST Student

### Endpoint

```http
POST /api/v1/students/
```

### Contoh Body

```json
{
  "nim": "434241099",
  "name": "Dina",
  "grade": 88.5,
  "is_active": true
}
```

### Status

- `201 Created` — student berhasil ditambahkan.
- `409 Conflict` — NIM sudah digunakan.
- `415 Unsupported Media Type` — Content-Type bukan JSON.
- `422 Unprocessable Entity` — data gagal validasi.
- `500 Internal Server Error` — terjadi kesalahan server atau database.

Jika berhasil, response memiliki header:

```text
Location: /api/v1/students/4
```

---

## 4. PUT Student

### Endpoint

```http
PUT /api/v1/students/:id
```

PUT digunakan untuk mengganti seluruh data student.

### Contoh Body

```json
{
  "nim": "434241101",
  "name": "Sania Putri",
  "grade": 95,
  "is_active": true
}
```

### Status

- `200 OK` — student berhasil diperbarui.
- `400 Bad Request` — ID atau request tidak valid.
- `404 Not Found` — student tidak ditemukan.
- `409 Conflict` — NIM menjadi duplikat.
- `415 Unsupported Media Type` — Content-Type bukan JSON.
- `422 Unprocessable Entity` — data gagal validasi.
- `500 Internal Server Error` — terjadi kesalahan server atau database.

---

## 5. PATCH Student

### Endpoint

```http
PATCH /api/v1/students/:id
```

PATCH digunakan untuk mengubah sebagian data student.

### Contoh Body

```json
{
  "grade": 88
}
```

### Status

- `200 OK` — student berhasil diperbarui.
- `400 Bad Request` — ID atau request tidak valid.
- `404 Not Found` — student tidak ditemukan.
- `409 Conflict` — NIM menjadi duplikat.
- `415 Unsupported Media Type` — Content-Type bukan JSON.
- `422 Unprocessable Entity` — data gagal validasi.
- `500 Internal Server Error` — terjadi kesalahan server atau database.

---

## 6. DELETE Student

### Endpoint

```http
DELETE /api/v1/students/:id
```

### Contoh Request

```http
DELETE /api/v1/students/3
```

### Status

- `204 No Content` — berhasil dihapus.
- `400 Bad Request` — ID tidak valid.
- `404 Not Found` — student tidak ditemukan.
- `500 Internal Server Error` — terjadi kesalahan server atau database.

Jika berhasil, response tidak memiliki body.

---

# Query String

## Pagination

```http
GET /api/v1/students/?page=1&limit=2
```

Nilai default:

```text
page = 1
limit = 10
```

Batas maksimum `limit` adalah `50`.

Batas maksimum 50 digunakan agar client tidak meminta terlalu banyak data dalam satu response sehingga penggunaan memori dan waktu pemrosesan server tetap terkendali.

---

## Search

```http
GET /api/v1/students/?search=sAnIa
```

Pencarian tidak membedakan huruf besar dan kecil karena menggunakan `ILIKE` pada PostgreSQL.

---

## Sorting

```http
GET /api/v1/students/?sort=grade&order=desc
```

Data akan diurutkan berdasarkan grade dari terbesar ke terkecil.

---

## Filter

Filter status aktif:

```http
GET /api/v1/students/?is_active=true
```

Filter berdasarkan rentang grade:

```http
GET /api/v1/students/?grade_min=80&grade_max=95
```

---

# Meta Pagination

Response daftar student memiliki:

```json
{
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 3,
    "total_pages": 1
  }
}
```

| Field | Keterangan |
|---|---|
| `page` | Halaman yang sedang ditampilkan |
| `limit` | Jumlah data per halaman |
| `total` | Jumlah seluruh data yang sesuai dengan filter |
| `total_pages` | Jumlah seluruh halaman |

---

# Pengujian Error

## 404 Not Found

Pengujian dilakukan dengan mengakses ID yang tidak terdapat di database.

Contoh:

```http
GET /api/v1/students/99999
```

Response:

```text
404 Not Found
```

Pengujian juga dilakukan pada:

```http
PUT /api/v1/students/99999
PATCH /api/v1/students/99999
DELETE /api/v1/students/99999
```

---

## 409 Conflict

Status `409 Conflict` diuji dengan memasukkan NIM yang sudah terdapat pada database.

Contoh:

```http
POST /api/v1/students/
```

Body:

```json
{
  "nim": "434241101",
  "name": "Mahasiswa Duplikat",
  "grade": 80,
  "is_active": true
}
```

Jika NIM `434241101` sudah terdapat pada database:

```text
409 Conflict
```

Hal tersebut terjadi karena kolom `nim` memiliki constraint:

```sql
UNIQUE
```

---

## 503 Service Unavailable

Untuk melakukan pengujian `503 Service Unavailable`, PostgreSQL dimatikan terlebih dahulu.

Setelah PostgreSQL dimatikan, API dipanggil kembali.

Contoh:

```http
GET /api/v1/students/
```

API akan gagal mengakses database dan mengembalikan:

```text
503 Service Unavailable
```

Status `503` dipilih karena server aplikasi masih berjalan, tetapi service database yang dibutuhkan untuk memproses request sedang tidak tersedia.

---

# Format Response

## Response Berhasil

```json
{
  "success": true,
  "message": "Pesan berhasil",
  "data": {}
}
```

## Response Error

```json
{
  "success": false,
  "message": "Pesan kesalahan",
  "error": {}
}
```

---

# Content-Type

Request berikut:

```text
POST
PUT
PATCH
```

harus menggunakan:

```http
Content-Type: application/json
```

Jika Content-Type tidak sesuai:

```text
415 Unsupported Media Type
```

---

# Menjalankan Project

Masuk ke folder project:

```powershell
cd D:\ProjectGo\api-students
```

Pastikan PostgreSQL sedang berjalan.

Pastikan file `.env` sudah tersedia dan konfigurasi database sudah benar.

Kemudian jalankan:

```powershell
go mod tidy
```

Setelah itu jalankan aplikasi:

```powershell
go run .
```

Server berjalan pada:

```text
http://localhost:3000
```

API:

```text
http://localhost:3000/api/v1
```

Health check:

```text
http://localhost:3000/health
```

---

# Contoh Request

## Health Check

```http
GET http://localhost:3000/health
```

## Mengambil Semua Student

```http
GET http://localhost:3000/api/v1/students/
```

## Mengambil Student Berdasarkan ID

```http
GET http://localhost:3000/api/v1/students/1
```

## Menambahkan Student

```http
POST http://localhost:3000/api/v1/students/
Content-Type: application/json
```

Body:

```json
{
  "nim": "434241099",
  "name": "Dina",
  "grade": 88.5,
  "is_active": true
}
```

## Mengubah Student

```http
PUT http://localhost:3000/api/v1/students/1
Content-Type: application/json
```

Body:

```json
{
  "nim": "434241101",
  "name": "Sania Putri",
  "grade": 95,
  "is_active": true
}
```

## Mengubah Sebagian Student

```http
PATCH http://localhost:3000/api/v1/students/1
Content-Type: application/json
```

Body:

```json
{
  "grade": 90
}
```

## Menghapus Student

```http
DELETE http://localhost:3000/api/v1/students/1
```

---

# Query Examples

## Pagination

```http
GET /api/v1/students/?page=1&limit=2
```

## Search

```http
GET /api/v1/students/?search=sania
```

## Sorting

```http
GET /api/v1/students/?sort=grade&order=desc
```

## Filter Status

```http
GET /api/v1/students/?is_active=true
```

## Filter Grade

```http
GET /api/v1/students/?grade_min=80&grade_max=95
```

## Kombinasi Query

```http
GET /api/v1/students/?page=1&limit=10&search=sania&sort=grade&order=desc&is_active=true&grade_min=80&grade_max=100
```

---

# Git dan Environment

File `.env` tidak boleh ikut di-commit ke GitHub.

Pastikan `.gitignore` memiliki:

```text
.env
```

File yang boleh di-commit adalah:

```text
.env.example
```

`.env.example` hanya berisi nama variabel environment dengan nilai kosong.

Contoh:

```env
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_SSLMODE=
```

Untuk memeriksa status Git:

```powershell
git status
```

Untuk memastikan `.env` tidak dilacak Git:

```powershell
git ls-files .env
```

Jika tidak ada output, berarti `.env` tidak sedang dilacak oleh Git.

---

# Commit Bertahap

Project dikerjakan menggunakan beberapa commit secara bertahap.

Contoh tahapan commit:

```text
feat: add students database schema
chore: add database environment configuration
feat: add PostgreSQL connection pool
feat: implement PostgreSQL student repository
feat: migrate student API to PostgreSQL
docs: update README
```

Commit bertahap digunakan agar setiap perubahan dapat dilihat dan dilacak dengan jelas melalui Git.

Untuk melihat riwayat commit:

```powershell
git log --oneline
```

Untuk mengirim commit ke GitHub:

```powershell
git push origin main
```

---

# Sumber Bantuan

Sumber bantuan yang digunakan:

- Dokumentasi Go.
- Dokumentasi Fiber v2.
- Dokumentasi PostgreSQL.
- Referensi REST API dan HTTP.
- ChatGPT sebagai alat bantu AI.

Bantuan AI digunakan untuk membantu memahami:

- Struktur project Go.
- Integrasi Go dengan PostgreSQL.
- PostgreSQL connection pool.
- Environment variable dan file `.env`.
- Database migration.
- Repository Pattern.
- Interface `StudentRepository`.
- Implementasi `FindAll`, `FindByID`, `Create`, `Update`, dan `Delete`.
- Sentinel error `ErrNotFound` dan `ErrDuplicate`.
- Parameterized query.
- SQL `ILIKE`.
- SQL `WHERE`.
- SQL `ORDER BY`.
- SQL `LIMIT` dan `OFFSET`.
- SQL `COUNT(*)`.
- Penanganan error database.
- Status HTTP `404`, `409`, dan `503`.
- Penyusunan dokumentasi README.
