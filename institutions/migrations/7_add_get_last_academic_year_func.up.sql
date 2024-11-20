-- Get last academic year
CREATE FUNCTION func_get_last_academic_year (IN institution BIGINT) RETURNS SETOF vw_AllAcademicYears LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY SELECT 
        *
    FROM
        vw_AllAcademicYears a
    WHERE
        a.institution_id=institution_id
    ORDER BY
        end_date DESC
    LIMIT 1
    ;
END
$$;