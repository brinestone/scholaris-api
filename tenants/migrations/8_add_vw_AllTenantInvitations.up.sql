CREATE VIEW
    vw_AllTenantInvitations AS
SELECT
    mi.id,
    mi."user",
    mi.tenant,
    t.name AS tenant_name,
    mi.email,
    mi.phone,
    mi."role",
    mi.display_name,
    mi.success_redirect,
    mi.error_redirect,
    mi.onboard_redirect,
    mi.url,
    mi.avatar,
    mi.created_at,
    mi.updated_at,
    (
        CASE
            WHEN NOW() < mi.updated_at::DATE + mi."window"
            AND tm.id IS NULL THEN 'pending'
            WHEN NOW() >= mi.updated_at::DATE + mi."window"
            OR tm.id IS NULL THEN 'expired'
            ELSE 'accepted'
        END
    )::invite_status AS invite_status,
    (
        CASE
            WHEN tm.id IS NULL THEN mi.updated_at::DATE + mi."window"
            ELSE NULL
        END
    ) AS expires_at
FROM
    member_invites mi
    JOIN tenants t ON t.id = mi.tenant
    LEFT JOIN tenant_memberships tm ON mi.id = tm.invite;