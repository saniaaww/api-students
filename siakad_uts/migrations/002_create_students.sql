CREATE TABLE students (
    id SERIAL PRIMARY KEY,

    user_id INTEGER NOT NULL UNIQUE,

    nim VARCHAR(12) NOT NULL UNIQUE,

    nama VARCHAR(100) NOT NULL,

    prodi VARCHAR(100) NOT NULL,

    angkatan INTEGER NOT NULL,

    ipk_terakhir NUMERIC(3,2)
        CHECK (
            ipk_terakhir IS NULL
            OR (
                ipk_terakhir >= 0
                AND ipk_terakhir <= 4
            )
        ),

    deleted_at TIMESTAMP NULL,

    CONSTRAINT fk_students_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
);