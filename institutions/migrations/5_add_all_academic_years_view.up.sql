-- All academic years view
CREATE VIEW
    vw_AllAcademicYears AS
SELECT
    ay.id AS year_id,
    ay.institution AS institution_id,
    (ay.created_at + ay.start_offset) AS start_date,
    EXTRACT(
        EPOCH
        FROM
            ay.duration
    )*1000000000 AS duration,
    (ay.created_at + ay.start_offset + ay.duration) AS end_date,
    (
        DATE_PART('year', ay.created_at)::TEXT || '/' || DATE_PART(
            'year',
            (ay.created_at + ay.start_offset + ay.duration)::DATE
        )
    ) AS label,
    ay.created_at AS created_at,
    ay.updated_at AS updated_at
FROM
    academic_years ay
ORDER BY
    end_date DESC;