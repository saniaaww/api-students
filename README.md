# API Students

REST API untuk mengelola data mahasiswa menggunakan Go dan Fiber v2.

Project ini dibuat untuk memenuhi tugas mandiri Pertemuan 3 — REST API & HTTP Deep Dive.

---

## Teknologi

- Go
- Fiber v2
- REST API
- JSON
- HTTP

---

## Struktur Project

```text
api-students/
├── main.go
├── model.go
├── helper.go
├── handler.go
├── go.mod
├── go.sum
└── README.md
```

### Penjelasan File

| File | Keterangan |
|---|---|
| `main.go` | Menjalankan server, middleware, dan routing API |
| `model.go` | Berisi struct Student dan struct request untuk POST, PUT, dan PATCH |
| `helper.go` | Berisi helper untuk membuat response berhasil dan response error |
| `handler.go` | Berisi data student dan handler untuk endpoint API |
| `README.md` | Dokumentasi kontrak API |

---

# Base URL

```text
http://localhost:3000/api/v1
```

---

# Struktur Data Student

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | int | ID unik student |
| `nim` | string | NIM unik mahasiswa |
| `name` | string | Nama mahasiswa |
| `grade` | float64 | Nilai mahasiswa |
| `is_active` | bool | Status aktif mahasiswa |

Contoh:

```json
{
  "id": 1,
  "nim": "434241101",
  "name": "Sania",
  "grade": 91.5,
  "is_active": true
}
```

---

# API Contract

| Metode | Endpoint | Parameter | Contoh Body Permintaan | Status yang Mungkin Dikembalikan | Contoh Respons |
|---|---|---|---|---|---|
| GET | `/api/v1/students/` | `page`, `limit`, `search`, `sort`, `order`, `is_active`, `grade_min`, `grade_max` | Tidak ada | `200`, `400` | `{"success":true,"message":"Data student berhasil diambil","data":[...],"meta":{...}}` |
| GET | `/api/v1/students/:id` | `id` | Tidak ada | `200`, `400`, `404` | `{"success":true,"message":"Student ditemukan","data":{...}}` |
| POST | `/api/v1/students/` | Tidak ada | `{"nim":"434241099","name":"Dina","grade":88.5,"is_active":true}` | `201`, `409`, `415`, `422` | `{"success":true,"message":"Student berhasil dibuat","data":{...}}` |
| PUT | `/api/v1/students/:id` | `id` | `{"nim":"434241101","name":"Sania Putri","grade":95,"is_active":true}` | `200`, `400`, `404`, `415`, `422` | `{"success":true,"message":"Student berhasil diperbarui","data":{...}}` |
| PATCH | `/api/v1/students/:id` | `id` | `{"grade":88}` | `200`, `400`, `404`, `415`, `422` | `{"success":true,"message":"Student berhasil diperbarui sebagian","data":{...}}` |
| DELETE | `/api/v1/students/:id` | `id` | Tidak ada | `204`, `400`, `404` | Tidak ada response body |

---

# Detail Endpoint

## 1. GET Daftar Student

### Endpoint

```http
GET /api/v1/students/
```

Digunakan untuk mengambil daftar seluruh student.

### Query Parameter

| Parameter | Keterangan | Default | Contoh |
|---|---|---|---|
| `page` | Nomor halaman | `1` | `page=1` |
| `limit` | Jumlah data per halaman | `10` | `limit=2` |
| `search` | Pencarian berdasarkan nama tanpa membedakan huruf besar dan kecil | - | `search=sania` |
| `sort` | Field yang digunakan untuk sorting | - | `sort=grade` |
| `order` | Urutan `asc` atau `desc` | - | `order=desc` |
| `is_active` | Filter berdasarkan status aktif | - | `is_active=true` |
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

Jika berhasil, response juga memiliki header:

```text
Location: /api/v1/students/4
```

---

## 4. PUT Student

### Endpoint

```http
PUT /api/v1/students/:id
```

PUT digunakan untuk **mengganti seluruh data** student.

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

- `200 OK`
- `400 Bad Request`
- `404 Not Found`
- `415 Unsupported Media Type`
- `422 Unprocessable Entity`

---

## 5. PATCH Student

### Endpoint

```http
PATCH /api/v1/students/:id
```

PATCH digunakan untuk **mengubah sebagian data** student.

### Contoh Body

```json
{
  "grade": 88
}
```

### Status

- `200 OK`
- `400 Bad Request`
- `404 Not Found`
- `415 Unsupported Media Type`
- `422 Unprocessable Entity`

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

Pencarian tidak membedakan huruf besar dan kecil.

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
| `total` | Jumlah seluruh data |
| `total_pages` | Jumlah seluruh halaman |

---

# Status HTTP

| Status | Keterangan |
|---|---|
| `200 OK` | Pengambilan atau perubahan data berhasil |
| `201 Created` | Data berhasil ditambahkan |
| `204 No Content` | Data berhasil dihapus tanpa response body |
| `400 Bad Request` | Request atau parameter tidak valid |
| `404 Not Found` | Data tidak ditemukan |
| `409 Conflict` | Terjadi konflik, misalnya NIM ganda |
| `415 Unsupported Media Type` | Content-Type bukan application/json |
| `422 Unprocessable Entity` | Isi request gagal validasi |

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

Request `POST`, `PUT`, dan `PATCH` harus menggunakan:

```http
Content-Type: application/json
```

Jika tidak sesuai:

```text
415 Unsupported Media Type
```

---

# Menjalankan Project

Masuk ke folder project:

```powershell
cd D:\ProjectGo\api-students
```

Jalankan:

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

---

# Sumber Bantuan

Sumber bantuan yang digunakan:

- Dokumentasi Go.
- Dokumentasi Fiber v2.
- Referensi REST API dan HTTP.
- ChatGPT sebagai alat bantu AI.

Bantuan AI digunakan untuk membantu memahami:

- Struktur project Go.
- Struct dan request struct.
- REST API menggunakan Fiber.
- Handler GET, POST, PUT, PATCH, dan DELETE.
- Pagination, search, sorting, dan filtering.
- Status HTTP.
- Penyusunan dokumentasi API.
