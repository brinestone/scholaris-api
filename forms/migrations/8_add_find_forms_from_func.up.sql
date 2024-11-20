CREATE FUNCTION func_find_forms (
    IN arg_owner_id BIGINT,
    IN arg_owner_type TEXT,
    IN arg_overrides BIGINT[],
    IN arg_page INT,
    IN arg_size INT
) RETURNS SETOF vw_AllForms LANGUAGE plpgsql AS $$
BEGIN
    IF arg_overrides IS NOT NULL AND ARRAY_LENGTH(arg_overrides, 1) > 0 THEN
        RETURN QUERY SELECT * FROM vw_AllForms f WHERE f.owner=arg_owner_id AND f.owner_type=arg_owner_type AND (f.status != 'draft' OR f.id=ANY(arg_overrides)) OFFSET arg_page * arg_size LIMIT arg_size;
    ELSE
        RETURN QUERY SELECT * FROM vw_AllForms f WHERE f.owner=arg_owner_id AND f.owner_type=arg_owner_type AND f.status != 'draft' OFFSET arg_page * arg_size LIMIT arg_size;
    END IF;
END;
$$;