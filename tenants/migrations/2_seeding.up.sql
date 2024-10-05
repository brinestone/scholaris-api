-- Database Seeding
BEGIN;
INSERT INTO subscription_plans (id, name, price) VALUES (1, 'Basic', 0);
INSERT INTO subscription_plans (id, name, price) VALUES (2, 'Medium', 50000);
INSERT INTO subscription_plans (id, name, price) VALUES (3, 'Ultimate', 100000);

-- Tenants count
INSERT INTO plan_benefits (id, name, subscription_plan, details, max_count, min_count) VALUES (1, '# of Organizations', 1, 'At most 1', 1, 0);
INSERT INTO plan_benefits (id, name, subscription_plan, details, max_count, min_count) VALUES (2, '# of Organizations', 2, 'At most 5', 5, 0);
INSERT INTO plan_benefits (id, name, subscription_plan, details, max_count, min_count) VALUES (3, '# of Organizations', 3, 'Unlimited', NULL, 0);
COMMIT;