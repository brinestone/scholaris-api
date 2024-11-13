-- All academic terms
CREATE VIEW
    vw_AllAcademicTerms AS
SELECT
    t.id,
    t.year_id,
    t.institution,
    (t.created_at + t.start_offset) AS start_date,
    EXTRACT(
        EPOCH
        FROM
            t.duration
    ) * 1000000000 as duration,
    (t.created_at + t.duration + t.start_offset) AS end_date,
    t.label,
    t.created_at,
    t.updated_at
FROM
    academic_terms t
GROUP BY
    t.id,
    t.institution
ORDER BY
    end_date DESC;