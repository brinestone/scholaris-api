CREATE FUNCTION func_create_internal_user (
    IN arg_email TEXT,
    IN arg_password_hash TEXT,
    IN arg_avatar TEXT,
    IN arg_first_name TEXT,
    IN arg_last_name TEXT,
    IN arg_gender TEXT,
    IN arg_dob DATE,
    IN arg_phone TEXT,
    IN arg_phone_verified BOOL,
    IN arg_email_verified BOOL
) RETURNS BIGINT AS $$
DECLARE
    user_id BIGINT;
    external_id TEXT;
    account_id BIGINT;
BEGIN
    INSERT INTO users(banned,locked) VALUES (DEFAULT, DEFAULT) RETURNING id INTO user_id;
    external_id := GEN_RANDOM_UUID()::TEXT;
    
    INSERT INTO provider_accounts("user", external_id,image_url,first_name,last_name,provider,password_hash,gender,dob) VALUES (
        user_id,external_id,arg_avatar,arg_first_name, arg_last_name,'internal',arg_password_hash,arg_gender,arg_dob
    ) RETURNING id INTO account_id;

    INSERT INTO account_emails(email, account, external_id, is_primary) VALUES(arg_email, account_id, GEN_RANDOM_UUID()::TEXT, COALESCE(arg_email_verified, false));
    INSERT INTO account_phones(phone, account, external_id, is_primary) VALUES(arg_phone, account_id, GEN_RANDOM_UUID()::TEXT, COALESCE(arg_phone_verified, false));

    RETURN user_id;
END;
$$ LANGUAGE plpgsql;