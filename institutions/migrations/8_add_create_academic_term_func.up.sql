-- Create academic term
CREATE FUNCTION func_create_new_academic_term (
    IN arg_year_id BIGINT,
    IN arg_institution_id BIGINT,
    IN arg_label TEXT,
    IN arg_duration INTERVAL,
    IN arg_start_offset INTERVAL
) RETURNS BIGINT LANGUAGE plpgsql AS $$
DECLARE
    actual_label TEXT;
    new_term_id BIGINT;
BEGIN
    actual_label := COALESCE(arg_label, 'New Term');
    INSERT INTO academic_terms(year_id, institution, duration, label, start_offset)
    VALUES(arg_year_id, arg_institution_id, arg_duration, actual_label, arg_start_offset)
    RETURNING id INTO new_term_id;

    RETURN new_term_id;
END
$$;