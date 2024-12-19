CREATE VIEW
    vw_AllTenantMembers AS
SELECT
    tm.id,
    COALESCE(tm.invite, mi.id) AS invite,
    mi."user",
    COALESCE(tm.display_name, mi.display_name) AS display_name,
    COALESCE(tm.avatar, mi.avatar) AS avatar,
    COALESCE(tm.email, mi.email) AS email,
    COALESCE(tm.phone, mi.phone) AS phone,
    tm.prefs,
    mi.tenant,
    mi.created_at AS invited_at,
    (
        CASE
            WHEN NOW() < mi.updated_at::DATE + mi."window"
            AND tm.id IS NULL THEN 'pending'
            WHEN NOW() >= mi.updated_at::DATE + mi."window"
            AND tm.id IS NULL THEN 'expired'
            ELSE 'accepted'
        END
    )::invite_status AS invite_status,
    (
        CASE
            WHEN tm.id IS NULL THEN mi.updated_at::DATE + mi."window"
            ELSE NULL
        END
    )::DATE AS invite_expires_at,
    tm.created_at,
    tm.updated_at,
    COALESCE(tm."role", mi."role") AS "role"
FROM
    member_invites mi
    LEFT JOIN tenant_memberships tm ON mi.id = tm.invite;