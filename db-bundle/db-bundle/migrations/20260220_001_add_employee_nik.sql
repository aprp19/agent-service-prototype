-- Transaction: TRUE
ALTER TABLE hris.employees
ADD COLUMN IF NOT EXISTS nik TEXT;

ALTER TABLE hris.employees
ADD CONSTRAINT employees_nik_unique UNIQUE (nik);
