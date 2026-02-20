-- Transaction: TRUE
ALTER TABLE hris.employees
ADD COLUMN IF NOT EXISTS phone TEXT;
