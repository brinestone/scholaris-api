CREATE FUNCTION func_get_academic_term (
    IN arg_institution_id BIGINT,
    IN arg_target_date DATE
) RETURNS SETOF vw_AllAcademicTerms LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY SELECT
        *
    FROM
        vw_AllAcademicTerms t
    WHERE
        t.institution = arg_institution_id AND
        COALESCE(arg_target_date, NOW()::DATE) BETWEEN t.start_date AND t.end_date; 
END;
$$;