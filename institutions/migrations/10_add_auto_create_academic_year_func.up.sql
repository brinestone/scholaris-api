CREATE FUNCTION func_auto_create_academic_year (
    IN arg_creation_offset TEXT,
    IN arg_start_offset TEXT,
    IN arg_institution_id BIGINT,
    IN arg_duration TEXT,
    IN arg_term_durations TEXT[],
    IN arg_vacation_durations TEXT[]
) RETURNS TABLE (year_id BIGINT, term_ids BIGINT[]) LANGUAGE plpgsql AS $$
DECLARE
    end_date DATE;
BEGIN
    SELECT a.end_date::DATE INTO end_date FROM proc_get_last_academic_year(arg_institution_id);
    IF NOW() >= end_date + arg_creation_offset::INTERVAL THEN
        RETURN QUERY SELECT * FROM func_create_academic_year(arg_institution_id, arg_start_offset,arg_duration,arg_term_durations,arg_vacation_durations);
    ELSE
        RETURN;
    END IF;
END
$$;