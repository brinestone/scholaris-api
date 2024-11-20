CREATE PROCEDURE proc_delete_user (IN user_id BIGINT) AS $$
DECLARE
    rows_affected INT;
BEGIN
    DELETE FROM users WHERE id=user_id;
    GET DIAGNOSTICS rows_affected = ROW_COUNT;

    IF rows_affected = 0 THEN
        RAISE EXCEPTION 'User account not found: %', user_id;
    END IF;
END;
$$ LANGUAGE plpgsql;