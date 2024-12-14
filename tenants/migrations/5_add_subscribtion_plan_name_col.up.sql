CREATE OR REPLACE VIEW
    vw_AllTenants AS
SELECT
    t.id,
    t.name,
    t.subscription,
    t.created_at,
    t.updated_at,
    ts.suspended,
    ts.subscription_plan,
    ts.next_billing_cycle,
    ts.created_at as subscribed_at,
    ts.updated_at as subscription_updated_at,
    sp.name as subscription_plan_name
FROM
    tenants t
    JOIN tenant_subscriptions ts ON ts.id = t.subscription
    JOIN subscription_plans sp ON ts.subscription_plan = sp.id;