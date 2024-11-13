CREATE TABLE
    academic_years (
        id BIGSERIAL,
        institution BIGINT NOT NULL,
        duration INTERVAL NOT NULL,
        start_offset INTERVAL NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (id),
        FOREIGN KEY (institution) REFERENCES institutions (id) ON DELETE CASCADE
    );

CREATE TABLE
    academic_terms (
        id BIGSERIAL,
        year_id BIGINT NOT NULL,
        institution BIGINT NOT NULL,
        label TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        duration INTERVAL NOT NULL,
        start_offset INTERVAL NOT NULL,
        PRIMARY KEY (id),
        -- UNIQUE (year_id, institution),
        FOREIGN KEY (institution) REFERENCES institutions (id),
        FOREIGN KEY (year_id) REFERENCES academic_years (id) ON DELETE CASCADE
    );