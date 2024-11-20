CREATE FUNCTION func_get_owner_settings (
    IN setting_ids BIGINT[],
    IN owner_id BIGINT,
    IN owner_type TEXT,
    IN fetch_all_if_no_setting_ids BOOLEAN,
    IN include_system_generated BOOLEAN
) RETURNS SETOF vw_AllSettings LANGUAGE plpgsql AS $$
BEGIN
  IF (setting_ids IS NULL OR array_length(setting_ids, 1) = 0) AND COALESCE(fetch_all_if_no_setting_ids, false) THEN
    RETURN QUERY 
    SELECT * 
    FROM vw_AllSettings s
    WHERE s.owner = owner_id;
  ELSE
    RETURN QUERY 
    SELECT * 
    FROM vw_AllSettings s
    WHERE s.owner = owner_id AND s.system_generated = COALESCE(include_system_generated, false) AND s.id = ANY(setting_ids);
  END IF;
END;
$$;