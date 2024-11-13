-- Get academic term
CREATE PROCEDURE proc_get_academic_term (
    IN institution_id BIGINT,
    IN target_date TIMESTAMP
) LANGUAGE plpgsql AS $$
BEGIN
    SELECT
        *
    FROM
        vw_AllAcademicTerms t
    WHERE
        t.institution=institution_id AND 
        COALESCE(target_date, NOW()) BETWEEN ay.start_date AND ay.end_date;
END
$$;