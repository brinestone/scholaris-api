CREATE TABLE
    enrollment_forms (
        form BIGINT NOT NULL,
        institution BIGINT NOT NULL,
        level BIGINT,
        PRIMARY KEY(form,institution),
        FOREIGN KEY(level) REFERENCES levels(id) ON DELETE CASCADE
    );