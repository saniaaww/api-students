CREATE TABLE users (
    id SERIAL PRIMARY KEY,

    email VARCHAR(255) NOT NULL UNIQUE,

    password VARCHAR(255) NOT NULL,

    role VARCHAR(20) NOT NULL
        CHECK (role IN ('admin', 'mahasiswa')),

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);