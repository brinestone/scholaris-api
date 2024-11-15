CREATE TABLE
    levels (
        id BIGSERIAL,
        institution BIGINT,
        name TEXT NOT NULL,
        PRIMARY KEY (id),
        FOREIGN KEY (institution) REFERENCES institutions (id) ON DELETE CASCADE
    );