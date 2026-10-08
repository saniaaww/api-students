CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (
    email,
    password,
    role
)
VALUES (
    'admin@siakad.com',
    crypt('admin12345', gen_salt('bf')),
    'admin'
);

INSERT INTO users (
    email,
    password,
    role
)
VALUES
(
    'sania.warda@siakad.com',
    crypt('434241100000', gen_salt('bf')),
    'mahasiswa'
),
(
    'aldha.zahro@siakad.com',
    crypt('434241100001', gen_salt('bf')),
    'mahasiswa'
),
(
    'edgina.cybil@siakad.com',
    crypt('434241100002', gen_salt('bf')),
    'mahasiswa'
),
(
    'nazlatul.khoiriyah@siakad.com',
    crypt('434241100003', gen_salt('bf')),
    'mahasiswa'
),
(
    'grachya.ayuma@siakad.com',
    crypt('434241100004', gen_salt('bf')),
    'mahasiswa'
),
(
    'ghefira.ahyan@siakad.com',
    crypt('434241100005', gen_salt('bf')),
    'mahasiswa'
),
(
    'aura.rieska@siakad.com',
    crypt('434241100006', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa08@siakad.com',
    crypt('434241100007', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa09@siakad.com',
    crypt('434241100008', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa10@siakad.com',
    crypt('434241100009', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa11@siakad.com',
    crypt('434241100010', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa12@siakad.com',
    crypt('434241100011', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa13@siakad.com',
    crypt('434241100012', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa14@siakad.com',
    crypt('434241100013', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa15@siakad.com',
    crypt('434241100014', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa16@siakad.com',
    crypt('434241100015', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa17@siakad.com',
    crypt('434241100016', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa18@siakad.com',
    crypt('434241100017', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa19@siakad.com',
    crypt('434241100018', gen_salt('bf')),
    'mahasiswa'
),
(
    'mahasiswa20@siakad.com',
    crypt('434241100019', gen_salt('bf')),
    'mahasiswa'
);


-- =====================================================
-- DATA STUDENTS
-- =====================================================

INSERT INTO students (
    user_id,
    nim,
    nama,
    prodi,
    angkatan,
    ipk_terakhir
)
SELECT
    id,
    CASE email
        WHEN 'sania.warda@siakad.com' THEN '434241100000'
        WHEN 'aldha.zahro@siakad.com' THEN '434241100001'
        WHEN 'edgina.cybil@siakad.com' THEN '434241100002'
        WHEN 'nazlatul.khoiriyah@siakad.com' THEN '434241100003'
        WHEN 'grachya.ayuma@siakad.com' THEN '434241100004'
        WHEN 'ghefira.ahyan@siakad.com' THEN '434241100005'
        WHEN 'aura.rieska@siakad.com' THEN '434241100006'
        WHEN 'mahasiswa08@siakad.com' THEN '434241100007'
        WHEN 'mahasiswa09@siakad.com' THEN '434241100008'
        WHEN 'mahasiswa10@siakad.com' THEN '434241100009'
        WHEN 'mahasiswa11@siakad.com' THEN '434241100010'
        WHEN 'mahasiswa12@siakad.com' THEN '434241100011'
        WHEN 'mahasiswa13@siakad.com' THEN '434241100012'
        WHEN 'mahasiswa14@siakad.com' THEN '434241100013'
        WHEN 'mahasiswa15@siakad.com' THEN '434241100014'
        WHEN 'mahasiswa16@siakad.com' THEN '434241100015'
        WHEN 'mahasiswa17@siakad.com' THEN '434241100016'
        WHEN 'mahasiswa18@siakad.com' THEN '434241100017'
        WHEN 'mahasiswa19@siakad.com' THEN '434241100018'
        WHEN 'mahasiswa20@siakad.com' THEN '434241100019'
    END,

    CASE email
        WHEN 'sania.warda@siakad.com' THEN 'Sania Warda'
        WHEN 'aldha.zahro@siakad.com' THEN 'Aldha Zahro'
        WHEN 'edgina.cybil@siakad.com' THEN 'Edgina Cybil'
        WHEN 'nazlatul.khoiriyah@siakad.com' THEN 'Nazlatul Khoiriyah'
        WHEN 'grachya.ayuma@siakad.com' THEN 'Grachya Ayuma'
        WHEN 'ghefira.ahyan@siakad.com' THEN 'Ghefira Ahyan'
        WHEN 'aura.rieska@siakad.com' THEN 'Aura Rieska'
        WHEN 'mahasiswa08@siakad.com' THEN 'Mahasiswa Delapan'
        WHEN 'mahasiswa09@siakad.com' THEN 'Mahasiswa Sembilan'
        WHEN 'mahasiswa10@siakad.com' THEN 'Mahasiswa Sepuluh'
        WHEN 'mahasiswa11@siakad.com' THEN 'Mahasiswa Sebelas'
        WHEN 'mahasiswa12@siakad.com' THEN 'Mahasiswa Dua Belas'
        WHEN 'mahasiswa13@siakad.com' THEN 'Mahasiswa Tiga Belas'
        WHEN 'mahasiswa14@siakad.com' THEN 'Mahasiswa Empat Belas'
        WHEN 'mahasiswa15@siakad.com' THEN 'Mahasiswa Lima Belas'
        WHEN 'mahasiswa16@siakad.com' THEN 'Mahasiswa Enam Belas'
        WHEN 'mahasiswa17@siakad.com' THEN 'Mahasiswa Tujuh Belas'
        WHEN 'mahasiswa18@siakad.com' THEN 'Mahasiswa Delapan Belas'
        WHEN 'mahasiswa19@siakad.com' THEN 'Mahasiswa Sembilan Belas'
        WHEN 'mahasiswa20@siakad.com' THEN 'Mahasiswa Dua Puluh'
    END,

    'D4 Teknik Informatika',

    2024,

    CASE email
        WHEN 'sania.warda@siakad.com' THEN 3.50
        WHEN 'aldha.zahro@siakad.com' THEN 3.20
        WHEN 'edgina.cybil@siakad.com' THEN 2.90
        WHEN 'nazlatul.khoiriyah@siakad.com' THEN 3.75
        WHEN 'grachya.ayuma@siakad.com' THEN 2.70
        WHEN 'ghefira.ahyan@siakad.com' THEN 3.10
        WHEN 'aura.rieska@siakad.com' THEN 2.40
        WHEN 'mahasiswa08@siakad.com' THEN 3.60
        WHEN 'mahasiswa09@siakad.com' THEN 2.85
        WHEN 'mahasiswa10@siakad.com' THEN 3.30
        WHEN 'mahasiswa11@siakad.com' THEN 3.00
        WHEN 'mahasiswa12@siakad.com' THEN 2.95
        WHEN 'mahasiswa13@siakad.com' THEN 3.80
        WHEN 'mahasiswa14@siakad.com' THEN 2.45
        WHEN 'mahasiswa15@siakad.com' THEN 3.15
        WHEN 'mahasiswa16@siakad.com' THEN 3.40
        WHEN 'mahasiswa17@siakad.com' THEN 2.75
        WHEN 'mahasiswa18@siakad.com' THEN 3.65
        WHEN 'mahasiswa19@siakad.com' THEN 2.55
        WHEN 'mahasiswa20@siakad.com' THEN 3.90
    END

FROM users
WHERE role = 'mahasiswa'
ORDER BY id;


-- =====================================================
-- COURSES
-- =====================================================

INSERT INTO courses (
    kode_mk,
    nama_mk,
    sks,
    semester,
    kuota
)
VALUES
(
    'IF101',
    'Algoritma dan Pemrograman',
    3,
    1,
    40
),
(
    'IF102',
    'Basis Data',
    3,
    2,
    40
),
(
    'IF103',
    'Pemrograman Berorientasi Objek',
    3,
    3,
    40
),
(
    'IF104',
    'Jaringan Komputer',
    3,
    3,
    30
),
(
    'IF105',
    'Rekayasa Perangkat Lunak',
    3,
    4,
    35
),
(
    'IF106',
    'Pengembangan Web',
    3,
    4,
    40
),
(
    'IF107',
    'Keamanan Siber',
    3,
    5,
    30
),
(
    'IF108',
    'Kecerdasan Buatan',
    3,
    5,
    30
),
(
    'IF109',
    'Pemrograman Mobile',
    3,
    5,
    35
),
(
    'IF110',
    'Manajemen Proyek TI',
    2,
    6,
    40
);

-- ADMIN
-- Email    : admin@siakad.com
-- Password : admin12345
-- =====================================================
-- MAHASISWA
-- Password awal = NIM