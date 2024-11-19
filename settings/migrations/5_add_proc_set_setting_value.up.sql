CREATE PROCEDURE proc_upsert_setting_value (
    IN arg_owner BIGINT,
    IN arg_owner_type TEXT,
    IN arg_user BIGINT,
    IN arg_updates JSONB[]
) AS $$
DECLARE
    acting_user BIGINT;
    current_setting_id BIGINT;
    current_update_key TEXT;
    current_update_key_value_updates JSONB[];
    current_setting_update JSONB;
    current_update_key_value_update JSONB;
    current_update_key_value_value TEXT;
    current_update_key_value_index TEXT;
BEGIN
    acting_user := COALESCE(arg_user, 0);
    IF arg_owner_type IS NULL OR arg_owner_type = '' THEN
        RAISE EXCEPTION 'Invalid value for arg_owner_type';
    END IF;
    IF arg_owner IS NULL OR arg_owner = 0 THEN
        RAISE EXCEPTION 'Invalid value for arg_owner';
    END IF;

    FOREACH current_setting_update IN ARRAY arg_updates LOOP
        FOR current_update_key, current_update_key_value_updates IN SELECT key,value FROM JSONB_EACH_TEXT(sent_update) LOOP
            
            SELECT id INTO current_setting_id FROM settings WHERE key=current_update_key AND owner=arg_owner AND owner_type=arg_owner_type AND system_generated=false;

            IF current_setting_id IS NULL THEN
                RAISE EXCEPTION 'No setting could be found for key %', current_update_key;
            END IF;

            FOREACH current_update_key_value_update IN ARRAY current_update_key_value_updates LOOP
                FOR current_update_key_value_index,current_update_key_value_value IN SELECT value,index FROM JSONB_EACH_TEXT(current_update_key_value_update) LOOP
                    INSERT INTO 
                        setting_values (set_by,value,set_at,setting,value_index)
                    VALUES
                        (acting_user,current_update_key_value_value,DEFAULT,current_setting_id,current_update_key_value_index)
                    ON CONFLICT
                        (setting,value_index)
                    DO
                        UPDATE SET
                            value=current_update_key_value_value,
                            set_by=acting_user,
                            set_at=DEFAULT;
                END LOOP;
            END LOOP;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;