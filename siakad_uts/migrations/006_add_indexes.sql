-- =====================================================
-- INDEX UNTUK STUDENTS
-- =====================================================

CREATE INDEX IF NOT EXISTS idx_students_nama
    ON students(nama);

CREATE INDEX IF NOT EXISTS idx_students_prodi
    ON students(prodi);

CREATE INDEX IF NOT EXISTS idx_students_angkatan
    ON students(angkatan);

CREATE INDEX IF NOT EXISTS idx_students_deleted_at
    ON students(deleted_at);


-- =====================================================
-- INDEX UNTUK COURSES
-- =====================================================

CREATE INDEX IF NOT EXISTS idx_courses_semester
    ON courses(semester);


-- =====================================================
-- INDEX UNTUK ENROLLMENTS
-- =====================================================

CREATE INDEX IF NOT EXISTS idx_enrollments_student_year
    ON enrollments(student_id, tahun_akademik);

CREATE INDEX IF NOT EXISTS idx_enrollments_course_year
    ON enrollments(course_id, tahun_akademik);