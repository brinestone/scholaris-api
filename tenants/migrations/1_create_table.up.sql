CREATE TABLE
    subscription_plans (
        id BIGSERIAL PRIMARY KEY,
        name VARCHAR(200) NOT NULL,
        created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        price DECIMAL DEFAULT 0,
        currency VARCHAR(3) DEFAULT 'XAF',
        enabled BOOLEAN DEFAULT true,
        billing_cycle INT DEFAULT 30
    );

CREATE TABLE
    plan_benefits (
        id BIGSERIAL PRIMARY KEY,
        name VARCHAR(200) NOT NULL,
        subscription_plan BIGINT REFERENCES subscription_plans (id) ON DELETE SET NULL,
        details text,
        min_count int,
        max_count int
    );

CREATE TABLE
    tenant_subscriptions (
        id BIGSERIAL PRIMARY KEY,
        subscription_plan BIGINT REFERENCES subscription_plans (id),
        next_billing_cycle TIMESTAMP,
        suspended BOOLEAN DEFAULT false,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE
    tenants (
        id BIGSERIAL PRIMARY KEY,
        name VARCHAR(255) UNIQUE NOT NULL,
        subscription BIGINT NOT NULL REFERENCES tenant_subscriptions (id) ON DELETE NO ACTION,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );