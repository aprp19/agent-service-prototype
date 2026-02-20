-- Baseline for HRIS Prototype

CREATE SCHEMA IF NOT EXISTS hris;
CREATE SCHEMA IF NOT EXISTS hris_meta;

CREATE TABLE IF NOT EXISTS hris_meta.schema_migrations (
    version TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    checksum TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    execution_time_ms BIGINT,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error TEXT
);

CREATE TABLE IF NOT EXISTS hris.employees (
    id BIGSERIAL PRIMARY KEY,
    employee_code TEXT UNIQUE,
    name TEXT NOT NULL,
    email TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS hris.audit_logs (
    id BIGSERIAL PRIMARY KEY,
    entity TEXT NOT NULL,
    entity_id TEXT,
    action TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);
