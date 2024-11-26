CREATE FUNCTION func_create_external_user (
    IN arg_first_name TEXT,
    IN arg_last_name TEXT,
    IN arg_external_id TEXT,
    IN arg_provider_data JSONB,
    IN arg_emails JSONB[], -- Expecting the structure: {"emailAddress":"string","verified":boolean,"externalId": "string","isPrimary": boolean}
    IN arg_phones JSONB[], -- Expecting the structure: {"phoneNumber":"string","verified":boolean,"externalId": "string","isPrimary": boolean}
    IN arg_provider TEXT,
    IN arg_gender TEXT,
    IN arg_dob TEXT,
    IN arg_avatar TEXT
) RETURNS BIGINT LANGUAGE plpgsql AS $$
DECLARE
    user_id              BIGINT;
    account_id           BIGINT;
    current_email_info   JSONB;
    current_phone_info   JSONB;
    current__            TEXT;
    current__external_id TEXT;
    current__verified    BOOL;
    current__primary     BOOL;
BEGIN
    INSERT INTO users DEFAULT VALUES RETURNING id INTO user_id;

    INSERT INTO provider_accounts(
            "user", 
            external_id,
            image_url,
            first_name,
            last_name,
            provider,
            gender,
            dob
        ) VALUES (
            user_id,
            arg_external_id,
            arg_avatar,
            arg_first_name,
            arg_last_name,
            arg_provider,
            arg_gender,
            arg_dob::DATE
        ) RETURNING id INTO account_id;

    FOREACH current_email_info IN ARRAY arg_emails LOOP
        current__ := current_email_info ->> 'emailAddress';
        current__verified := current_email_info ->> 'verified';
        current__external_id := COALESCE(current_email_info ->> 'externalId', GEN_RANDOM_UUID()::TEXT);
        current__primary := current_email_info ->> 'isPrimary';

        INSERT INTO account_emails(
            email,
            account,
            external_id,
            is_primary,
            verified
        ) 
            VALUES
        (
            current__,
            account_id,
            current__external_id,
            current__primary,
            current__verified
        );
    END LOOP;

    FOREACH current_phone_info IN ARRAY arg_phones LOOP
        current__ := current_phone_info ->> 'phoneNumber';
        current__external_id := COALESCE(current_phone_info ->> 'externalId', GEN_RANDOM_UUID()::TEXT);
        current__primary := current_phone_info ->> 'isPrimary';

        INSERT INTO account_phones(
            phone,
            account,
            external_id,
            is_primary,
            verified
        ) 
            VALUES
        (
            current__,
            account_id,
            current__external_id,
            current__primary,
            current__verified
        );
    END LOOP;

    RETURN user_id;
END;
$$;