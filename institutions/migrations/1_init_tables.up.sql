CREATE TABLE
    institutions (
        id BIGSERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description VARCHAR(255) DEFAULT NULL,
        logo VARCHAR(255) DEFAULT NULL,
        visible BOOLEAN DEFAULT FALSE,
        slug VARCHAR(15) NOT NULL,
        tenant BIGINT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        verified BOOLEAN DEFAULT false
    );

CREATE UNIQUE INDEX IDX_UQ_tenant_slug_1 ON institutions (tenant, slug);

CREATE TABLE
    enrollment_forms (
        id BIGINT NOT NULL,
        form BIGINT NOT NULL,
        institution BIGINT NOT NULL,
        FOREIGN KEY (institution) REFERENCES institutions (id)
    );