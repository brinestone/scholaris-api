-- Create academic year
CREATE FUNCTION func_create_academic_year (
    IN arg_institution_id BIGINT,
    IN arg_start_offset TEXT,
    IN arg_term_durations TEXT[],
    IN arg_vacation_durations TEXT[]
) RETURNS TABLE (year_id BIGINT, term_ids BIGINT[]) LANGUAGE plpgsql AS $$
DECLARE
    term_duration INTERVAL;
    term_ids BIGINT[];
    term_id BIGINT;
    year_id BIGINT;
    computed_term_start_offset INTERVAL;
    computed_term_label TEXT;
    computed_year_duration INTERVAL;
    i INT;
    j INT;
BEGIN

    IF ARRAY_LENGTH(arg_vacation_durations,1) > 0 AND ARRAY_LENGTH(arg_vacation_durations,1) != ARRAY_LENGTH(arg_term_durations,1)-1 THEN
        RAISE EXCEPTION 'The number of vacations should be 1 less than the number of terms. Expected %, got %', ARRAY_LENGTH(arg_term_durations,1)-1, ARRAY_LENGTH(arg_vacation_durations,1);
    END IF;

    computed_year_duration := '0 seconds'::INTERVAL;

    FOR i IN 1..ARRAY_LENGTH(arg_term_durations,1) LOOP
        computed_year_duration := computed_year_duration + arg_term_durations[i]::INTERVAL;
    END LOOP;
    FOR i IN 1..ARRAY_LENGTH(arg_vacation_durations,1) LOOP
        computed_year_duration := computed_year_duration + arg_vacation_durations[i]::INTERVAL;
    END LOOP;

    INSERT INTO academic_years(institution,duration,start_offset)
    VALUES(arg_institution_id,computed_year_duration,arg_start_offset::INTERVAL) RETURNING id INTO year_id;

    FOR i IN 1..ARRAY_LENGTH(arg_term_durations, 1) LOOP
        computed_term_label := 'Term' || ' ' || i::TEXT;
        IF i = 1 THEN
            computed_term_start_offset := COALESCE(arg_start_offset::INTERVAL, '0 seconds'::INTERVAL);
        ELSE 
            FOR j in 1..i-1 LOOP
                computed_term_start_offset := COALESCE(computed_term_start_offset, '0 seconds'::INTERVAL) + COALESCE(arg_vacation_durations[GREATEST(j-1,1)]::INTERVAL, '0 seconds'::INTERVAL) + COALESCE(arg_term_durations[j]::INTERVAL, '0 seconds'::INTERVAL);
            END LOOP;
        END IF;
        SELECT func_create_new_academic_term(year_id,arg_institution_id,computed_term_label,arg_term_durations[i]::INTERVAL,computed_term_start_offset) INTO term_id;
        IF term_id IS NULL THEN
            RAISE EXCEPTION 'Term ID is null for term_ids[%]',i;
        END IF;
        term_ids := ARRAY_APPEND(term_ids, term_id);
        computed_term_start_offset := COALESCE(arg_start_offset::INTERVAL, '0 seconds'::INTERVAL); -- reset this variable to default offset.
    END LOOP;
    RETURN QUERY SELECT year_id,term_ids;
END
$$;