CREATE
OR REPLACE FUNCTION calculate_next_value_index () RETURNS TRIGGER AS $$
DECLARE
    max_index INTEGER;
BEGIN
    IF TG_OP = 'INSERT' THEN
        SELECT MAX(value_index) INTO max_index FROM setting_values WHERE setting=NEW.setting;
        NEW.value_index := COALESCE(max_index, -1) + 1;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_value_index_trigger BEFORE INSERT ON setting_values FOR EACH ROW
EXECUTE FUNCTION calculate_next_value_index ();