CREATE TABLE courses (
    id SERIAL PRIMARY KEY,

    kode_mk VARCHAR(20) NOT NULL UNIQUE,

    nama_mk VARCHAR(150) NOT NULL,

    sks INTEGER NOT NULL
        CHECK (sks > 0),

    semester INTEGER NOT NULL
        CHECK (semester > 0),

    kuota INTEGER NOT NULL
        CHECK (kuota >= 0)
);