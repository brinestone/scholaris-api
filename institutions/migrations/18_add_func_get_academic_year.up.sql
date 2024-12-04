CREATE FUNCTION func_get_academic_year(
    IN arg_institution_id BIGINT,
    IN arg_target_date DATE
) RETURNS SETOF vw_AllAcademicYears LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY SELECT
        *
    FROM
        vw_AllAcademicYears y
    WHERE
        y.institution_id=arg_institution_id AND COALESCE(arg_target_date, NOW()::DATE) BETWEEN y.start_date AND y.end_date;
END;
$$;