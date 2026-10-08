CREATE TABLE enrollments (
    id SERIAL PRIMARY KEY,

    student_id INTEGER NOT NULL,

    course_id INTEGER NOT NULL,

    tahun_akademik VARCHAR(30) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_enrollments_student
        FOREIGN KEY (student_id)
        REFERENCES students(id),

    CONSTRAINT fk_enrollments_course
        FOREIGN KEY (course_id)
        REFERENCES courses(id),

    CONSTRAINT unique_student_course_year
        UNIQUE (
            student_id,
            course_id,
            tahun_akademik
        )
);

CREATE INDEX idx_enrollments_student
    ON enrollments(student_id);

CREATE INDEX idx_enrollments_course
    ON enrollments(course_id);