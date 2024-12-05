CREATE TABLE
    uploads (
        name TEXT NOT NULL,
        mime_type TEXT NOT NULL,
        size BIGINT DEFAULT 0,
        uploaded_by BIGINT NOT NULL,
        owner BIGINT,
        owner_type TEXT,
        key TEXT NOT NULL,
        uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (key)
    );

CREATE TABLE
    downloads (
        downloaded_by BIGINT,
        downloaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        key TEXT NOT NULL,
        FOREIGN KEY (key) REFERENCES uploads (key) ON DELETE CASCADE
    );