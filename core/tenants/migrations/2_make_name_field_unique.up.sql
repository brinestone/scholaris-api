ALTER TABLE tenants
    ADD CONSTRAINT unique_tenant_name_k UNIQUE (name);