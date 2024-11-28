CREATE FUNCTION func_find_user_password_hash (IN arg_user_id TEXT) RETURNS TEXT LANGUAGE plpgsql AS $$
DECLARE
    "password" TEXT;
BEGIN
    SELECT password_hash INTO "password" FROM provider_accounts WHERE "user"= arg_user_id;
    RETURN "password";
END;
$$;