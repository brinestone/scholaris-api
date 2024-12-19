CREATE TYPE invite_status AS ENUM('pending', 'accepted', 'expired');

CREATE TABLE
    member_invites (
        id BIGSERIAL,
        "user" BIGINT,
        tenant BIGINT NOT NULL,
        email VARCHAR(50) NOT NULL,
        phone VARCHAR(50),
        "role" VARCHAR(20) NOT NULL,
        display_name VARCHAR(100),
        success_redirect VARCHAR(500),
        error_redirect VARCHAR(500),
        onboard_redirect VARCHAR(500),
        url VARCHAR(500),
        avatar TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "window" INTERVAL DEFAULT '7 days'::INTERVAL,
        PRIMARY KEY (id),
        UNIQUE ("user", tenant),
        FOREIGN KEY (tenant) REFERENCES tenants (id) ON DELETE CASCADE
    );

CREATE TABLE
    tenant_memberships (
        id BIGSERIAL,
        invite BIGINT NOT NULL,
        avatar TEXT,
        "role" VARCHAR(20) NOT NULL,
        email VARCHAR(50) NOT NULL,
        phone VARCHAR(50),
        display_name VARCHAR(100) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        prefs JSONB,
        PRIMARY KEY (id),
        UNIQUE (invite),
        FOREIGN KEY (invite) REFERENCES member_invites (id) ON DELETE CASCADE
    );