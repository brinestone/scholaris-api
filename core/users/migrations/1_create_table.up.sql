DROP TABLE IF EXISTS users;

CREATE TYPE provider_type AS ENUM('internal', 'clerk');

CREATE TABLE
    users (
        id BIGSERIAL,
        banned BOOL DEFAULT false,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        locked BOOL DEFAULT false,
        PRIMARY KEY (id)
    );

CREATE TABLE
    provider_accounts (
        id BIGSERIAL,
        "user" BIGINT NOT NULL,
        external_id TEXT NOT NULL,
        image_url TEXT,
        first_name TEXT,
        last_name TEXT,
        provider provider_type NOT NULL DEFAULT 'internal',
        password_hash TEXT,
        provider_profile_data JSONB,
        gender TEXT,
        dob DATE,
        PRIMARY KEY (id),
        UNIQUE (external_id),
        UNIQUE ("user", external_id, provider),
        FOREIGN KEY ("user") REFERENCES users (id)
    );

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE
    account_emails (
        id BIGSERIAL,
        email TEXT NOT NULL,
        account BIGINT NOT NULL,
        external_id TEXT NOT NULL,
        is_primary BOOL DEFAULT false,
        verified BOOL DEFAULT false,
        PRIMARY KEY (id),
        UNIQUE (email),
        FOREIGN KEY (account) REFERENCES provider_accounts (id) ON DELETE CASCADE
    );

CREATE TABLE
    account_phones (
        id BIGSERIAL,
        phone TEXT NOT NULL,
        account BIGINT NOT NULL,
        external_id TEXT NOT NULL,
        is_primary BOOL DEFAULT false,
        verified BOOL DEFAULT false,
        PRIMARY KEY (id),
        UNIQUE (phone),
        FOREIGN KEY (account) REFERENCES provider_accounts (id) ON DELETE CASCADE
    );