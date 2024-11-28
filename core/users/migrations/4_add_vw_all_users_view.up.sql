CREATE VIEW
    vw_AllUsers AS
SELECT
    u.id,
    u.banned,
    u.created_at,
    u.updated_at,
    u.locked,
    COALESCE(
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'Id',
                pa.id,
                'ExternalId',
                pa.external_id,
                'ImageUrl',
                pa.image_url,
                'FirstName',
                pa.first_name,
                'LastName',
                pa.last_name,
                'Provider',
                pa.provider::TEXT,
                'ProviderProfileData',
                pa.provider_profile_data,
                'Gender',
                pa.gender,
                'Dob',
                pa.dob,
                'User',
                pa."user"
            )
        ) FILTER (
            WHERE
                pa.user IS NOT NULL
        ),
        '[]'
    ) AS providedAccounts,
    (
        SELECT
            id
        FROM
            account_emails _ae
        WHERE
            _ae.is_primary = true
            AND _ae.account = pa.id
    ) AS primary_email,
    (
        SELECT
            id
        FROM
            account_phones _ap
        WHERE
            _ap.is_primary = true
            AND _ap.account = pa.id
    ) AS primary_phone,
    COALESCE(
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'Id',
                ae.id,
                'Email',
                ae.email,
                'Account',
                ae.account,
                'ExternalId',
                ae.external_id,
                'IsPrimary',
                ae.is_primary,
                'Verified',
                ae.verified
            )
        ) FILTER (
            WHERE
                ae.account IS NOT NULL
        ),
        '[]'
    ) AS email_addresses,
    COALESCE(
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'Id',
                ap.id,
                'Phone',
                ap.phone,
                'Account',
                ap.account,
                'ExternalId',
                ap.external_id,
                'IsPrimary',
                ap.is_primary,
                'Verified',
                ap.verified
            )
        )
    ) AS phone_numbers
FROM
    users u
    LEFT JOIN provider_accounts pa ON pa."user" = u.id
    LEFT JOIN account_emails ae ON ae.account = pa.id
    LEFT JOIN account_phones ap ON ap.account = pa.id
GROUP BY
    u.id,
    pa.id;