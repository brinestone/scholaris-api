-- Get academic term by date
CREATE PROCEDURE proc_get_academic_year (
    IN institution_id BIGINT,
    IN target_date TIMESTAMP
) LANGUAGE plpgsql AS $$
BEGIN
    SELECT
        *
    FROM
        vw_AllAcademicYears ay
    WHERE
        ay.institution_id=institution_id AND COALESCE(target_date, NOW()) BETWEEN ay.start_date AND ay.end_date
    -- GROUP BY
    --     ay.year_id
    LIMIT 1;
END
$$;