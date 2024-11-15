CREATE TABLE
    enrollment_sessions (
        id BIGSERIAL,
        year_id BIGINT NOT NULL,
        enrollment BIGINT NOT NULL,
        owner BIGINT NOT NULL,
        level BIGINT NOT NULL,
        PRIMARY KEY (id),
        FOREIGN KEY (level) REFERENCES levels (id),
        FOREIGN KEY (year_id) REFERENCES academic_years (id),
        FOREIGN KEY (enrollment) REFERENCES enrollments (id)
    );