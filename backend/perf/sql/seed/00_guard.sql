\set ON_ERROR_STOP ON

DO $guard$
BEGIN
    IF current_database() <> 'coffeesys_perf' THEN
        RAISE EXCEPTION
            'refusing performance seed: current database is %, expected coffeesys_perf',
            current_database();
    END IF;
END
$guard$;

