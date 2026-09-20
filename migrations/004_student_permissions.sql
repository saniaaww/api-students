-- 1. Tambahkan permission untuk entity students
INSERT INTO permissions (name, description) VALUES
    ('student:list', 'Melihat daftar mahasiswa'),
    ('student:read:any', 'Melihat data mahasiswa mana pun'),
    ('student:create', 'Mendaftarkan data mahasiswa'),
    ('student:update:any', 'Mengubah data mahasiswa mana pun'),
    ('student:delete', 'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

-- 2. Pasangkan permission dengan role

-- ADMIN
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete')
ON CONFLICT DO NOTHING;

-- STAFF
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;

-- 3. Tambahkan owner_id ke students
ALTER TABLE students
ADD COLUMN IF NOT EXISTS owner_id BIGINT;

-- 4. Data students yang sudah ada diberikan kepada user id = 1
UPDATE students
SET owner_id = 1
WHERE owner_id IS NULL;

-- 5. Hubungkan owner_id dengan users.id
ALTER TABLE students
DROP CONSTRAINT IF EXISTS students_owner_id_fkey;

ALTER TABLE students
ADD CONSTRAINT students_owner_id_fkey
FOREIGN KEY (owner_id)
REFERENCES users(id)
ON DELETE RESTRICT;

-- 6. owner_id tidak boleh kosong
ALTER TABLE students
ALTER COLUMN owner_id SET NOT NULL;

-- 7. Index owner_id
CREATE INDEX IF NOT EXISTS idx_students_owner_id
ON students(owner_id);