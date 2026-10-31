
CREATE TABLE supervisors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    quota INTEGER NOT NULL CHECK (quota >= 0),
    supervised_count INTEGER NOT NULL DEFAULT 0
        CHECK (supervised_count >= 0),
    version INTEGER NOT NULL DEFAULT 1
        CHECK (version >= 1),
    CONSTRAINT chk_supervisor_quota
        CHECK (supervised_count <= quota)
);

CREATE TABLE supervisions (
    student_id INTEGER PRIMARY KEY
        REFERENCES students(id) ON DELETE CASCADE,
    supervisor_id INTEGER NOT NULL
        REFERENCES supervisors(id) ON DELETE RESTRICT
);

CREATE INDEX idx_supervisions_supervisor_id
    ON supervisions(supervisor_id);
