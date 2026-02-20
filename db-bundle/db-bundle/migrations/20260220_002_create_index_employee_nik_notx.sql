-- Transaction: FALSE
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_employees_nik
ON hris.employees (nik);
