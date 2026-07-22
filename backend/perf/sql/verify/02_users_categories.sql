\set ON_ERROR_STOP on

\ir ../seed/00_guard.sql
\ir ../seed/00_profile.sql


DO $verify$
DECLARE
    actual_users BIGINT;
    actual_categories BIGINT;
    invalid_users BIGINT;
    invalid_categories BIGINT;
    duplicate_emails BIGINT;
    duplicate_category_names BIGINT;
    users_sequence_value BIGINT;
    users_sequence_called BOOLEAN;
    categories_sequence_value BIGINT;
    expected_users BIGINT;
    categories_sequence_called BOOLEAN;
BEGIN
    SELECT users_count
    INTO expected_users
    FROM pg_temp.perf_profile;

    SELECT count(*) INTO actual_users
    FROM public.users;

    IF actual_users <> expected_users THEN
        RAISE EXCEPTION
            'users count mismatch: expected %, actual %',
            expected_users,
            actual_users;
    END IF;

    SELECT count(*) INTO actual_categories
    FROM public.categories;

    IF actual_categories <> 10 THEN
        RAISE EXCEPTION
            'categories count mismatch: expected 10, actual %',
            actual_categories;
    END IF;

    SELECT count(*) INTO invalid_users
    FROM public.users
    WHERE id NOT BETWEEN 1 AND expected_users
       OR name <> format('user-%s', lpad(id::text, 6, '0'))
       OR email <> format('user-%s@example.test', lpad(id::text, 6, '0'))
       OR password_hash <> 'perf-not-a-real-password-hash'
       OR role <> 'member'
       OR status <> 'active'
       OR reset_token IS NOT NULL
       OR created_at <> TIMESTAMPTZ '2025-01-01 00:00:00+00'
       OR updated_at <> TIMESTAMPTZ '2025-01-01 00:00:00+00';

    IF invalid_users <> 0 THEN
        RAISE EXCEPTION
            'invalid users found: %',
            invalid_users;
    END IF;

    SELECT count(*) - count(DISTINCT email)
    INTO duplicate_emails
    FROM public.users;

    IF duplicate_emails <> 0 THEN
        RAISE EXCEPTION
            'duplicate user emails found: %',
            duplicate_emails;
    END IF;

    SELECT count(*) INTO invalid_categories
    FROM public.categories
    WHERE id NOT BETWEEN 1 AND 10
       OR name <> format('category-%s', lpad(id::text, 2, '0'))
       OR created_at <> TIMESTAMPTZ '2025-01-01 00:00:00+00'
       OR updated_at <> TIMESTAMPTZ '2025-01-01 00:00:00+00';

    IF invalid_categories <> 0 THEN
        RAISE EXCEPTION
            'invalid categories found: %',
            invalid_categories;
    END IF;

    SELECT count(*) - count(DISTINCT name)
    INTO duplicate_category_names
    FROM public.categories;

    IF duplicate_category_names <> 0 THEN
        RAISE EXCEPTION
            'duplicate category names found: %',
            duplicate_category_names;
    END IF;

    SELECT last_value, is_called
    INTO users_sequence_value, users_sequence_called
    FROM public.users_id_seq;

    IF users_sequence_value <> expected_users OR NOT users_sequence_called THEN
        RAISE EXCEPTION
            'users sequence mismatch: expected value %, actual value %, called %',
            expected_users,
            users_sequence_value,
            users_sequence_called;
    END IF;

    SELECT last_value, is_called
    INTO categories_sequence_value, categories_sequence_called
    FROM public.categories_id_seq;

    IF categories_sequence_value <> 10 OR NOT categories_sequence_called THEN
        RAISE EXCEPTION
            'categories sequence mismatch: value %, called %',
            categories_sequence_value,
            categories_sequence_called;
    END IF;
END
$verify$;

\echo 'PASS: users and categories'