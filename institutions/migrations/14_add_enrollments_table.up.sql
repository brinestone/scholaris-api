CREATE TABLE
    enrollments (
        id BIGINT NOT NULL,
        form BIGINT NOT NULL,
        institution BIGINT NOT NULL,
        responder BIGINT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (id),
        FOREIGN KEY (form, institution) REFERENCES enrollment_forms (form, institution),
        FOREIGN KEY (institution) REFERENCES institutions (id) ON DELETE CASCADE
    );