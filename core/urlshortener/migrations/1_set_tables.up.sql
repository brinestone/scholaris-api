CREATE TABLE
    links (
        "key" VARCHAR(6),
        click_count INTEGER DEFAULT 0,
        max_clicks INTEGER,
        error_redirect VARCHAR(500),
        success_redirect VARCHAR(500) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "window" INTERVAL,
        PRIMARY KEY ("key")
    );

CREATE VIEW
    vw_AllLinks AS
SELECT
    l."key",
    l.click_count,
    l.max_clicks,
    l.error_redirect,
    l.success_redirect,
    l.created_at,
    l.updated_at,
    (
        CASE
            WHEN l."window" IS NOT NULL THEN l.created_at::DATE + l."window"
            ELSE NULL
        END
    ) AS expires_at,
    (
        (
            l."window" IS NULL
            OR NOW() <= (l.created_at::DATE + l."window")
        )
        AND (
            l.max_clicks IS NULL
            AND l.click_count < l.max_clicks
        )
    ) AS is_usable
FROM
    links l;