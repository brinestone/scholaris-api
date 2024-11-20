-- Institution statistics
CREATE VIEW
    vw_InstitutionStatistics AS
SELECT
    COUNT(id) AS total,
    (
        SELECT
            COUNT(id)
        FROM
            institutions
        WHERE
            verified = true
    ) AS verified,
    (
        SELECT
            COUNT(id)
        FROM
            institutions
        WHERE
            verified = false
    ) AS unverified
FROM
    institutions;