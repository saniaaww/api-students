CREATE TABLE students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE prestasi (
    idprestasi SERIAL PRIMARY KEY,
    id SERIAL FOREIGN KEY student, 
    name VARCHAR(100) NOT NULL,
    Juara VARCHAR (1) NOT NULL
); 
	
CREATE INDEX idx_students_name ON students(name);