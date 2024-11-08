CREATE TABLE
    institution_settings (
        id BIGSERIAL,
        institution BIGINT NOT NULL,
        system_generated BOOlEAN DEFAULT true,
        key TEXT NOT NULL,
        label TEXT NOT NULL,
        description TEXT,
        multi_value BOOLEAN DEFAULT false,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_by BIGINT DEFAULT 0,
        is_required BOOLEAN DEFAULT false,
        parent BIGINT,
        parent_type TEXT,
        PRIMARY KEY (id),
        UNIQUE (institution, key),
        FOREIGN KEY (institution) REFERENCES institutions (id) ON DELETE CASCADE
    );

CREATE TABLE
    setting_values (
        id BIGSERIAL,
        setting BIGINT NOT NULL,
        value TEXT NOT NULL,
        set_by BIGINT DEFAULT 0,
        set_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (id),
        FOREIGN KEY (setting) REFERENCES institution_settings (id) ON DELETE CASCADE
    );