SELECT 1 FROM hris_meta.schema_migrations LIMIT 1;
SELECT 1 FROM hris.employees LIMIT 1;
SELECT nik FROM hris.employees LIMIT 1;

SELECT indexname
FROM pg_indexes
WHERE schemaname = 'hris'
AND indexname = 'idx_employees_nik';
