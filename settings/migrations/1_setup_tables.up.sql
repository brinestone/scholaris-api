CREATE TABLE
    settings (
        id BIGSERIAL,
        label TEXT NOT NULL,
        description TEXT,
        key TEXT NOT NULL,
        multi_values BOOLEAN NOT NULL DEFAULT false,
        system_generated BOOLEAN NOT NULL DEFAULT false,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        parent BIGINT,
        owner BIGINT NOT NULL,
        owner_type TEXT NOT NULL,
        created_by BIGINT NOT NULL,
        overridable BOOLEAN DEFAULT true,
        PRIMARY KEY (id),
        UNIQUE (owner, owner_Type, key)
    );

CREATE TABLE
    setting_options (
        id BIGSERIAL,
        label TEXT NOT NULL,
        value TEXT,
        setting BIGINT NOT NULL,
        PRIMARY KEY (id),
        FOREIGN KEY (setting) REFERENCES settings (id) ON DELETE CASCADE
    );

CREATE TABLE
    setting_values (
        id BIGSERIAL,
        setting BIGINT NOT NULL,
        value TEXT,
        set_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        set_by BIGINT NOT NULL,
        PRIMARY KEY (id),
        FOREIGN KEY (setting) REFERENCES settings (id) ON DELETE CASCADE
    );