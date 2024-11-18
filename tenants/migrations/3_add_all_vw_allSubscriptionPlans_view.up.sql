CREATE VIEW
    vw_AllSubscriptionPlans AS
SELECT
    sp.id,
    sp.name,
    sp.created_at,
    sp.updated_at,
    sp.price,
    sp.currency,
    sp.enabled,
    sp.billing_cycle AS "billingCycle",
    COALESCE(
        json_agg(
            json_build_object(
                'Name',
                spd.name,
                'Details',
                spd.details,
                'MaxCount',
                spd.max_count,
                'MinCount',
                spd.min_count
            )
        ) FILTER (
            WHERE
                spd.id IS NOT NULL
        ),
        '[]'
    ) AS "descriptionsJson"
FROM
    subscription_plans sp
    LEFT JOIN plan_benefits spd ON sp.id = spd.subscription_plan
GROUP BY
    sp.id
ORDER BY
    sp.price ASC;