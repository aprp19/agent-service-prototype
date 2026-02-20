#!/usr/bin/env bash
set -e

OUT_DIR="db-bundle"
ZIP_NAME="db-bundle.zip"

echo "==> Creating bundle structure..."
rm -rf $OUT_DIR
mkdir -p $OUT_DIR/baseline
mkdir -p $OUT_DIR/migrations
mkdir -p $OUT_DIR/checks

########################################
# 1. BASELINE SQL
########################################
cat > $OUT_DIR/baseline/20260220_000_baseline.sql << 'EOF'
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
EOF

########################################
# 2. MIGRATION 001 (TX)
########################################
cat > $OUT_DIR/migrations/20260220_001_add_employee_nik.sql << 'EOF'
-- Transaction: TRUE
ALTER TABLE hris.employees
ADD COLUMN IF NOT EXISTS nik TEXT;

ALTER TABLE hris.employees
ADD CONSTRAINT employees_nik_unique UNIQUE (nik);
EOF

########################################
# 3. MIGRATION 002 (NON-TX)
########################################
cat > $OUT_DIR/migrations/20260220_002_create_index_employee_nik_notx.sql << 'EOF'
-- Transaction: FALSE
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_employees_nik
ON hris.employees (nik);
EOF

########################################
# 4. SMOKE CHECK
########################################
cat > $OUT_DIR/checks/smoke.sql << 'EOF'
SELECT 1 FROM hris_meta.schema_migrations LIMIT 1;
SELECT 1 FROM hris.employees LIMIT 1;
SELECT nik FROM hris.employees LIMIT 1;

SELECT indexname
FROM pg_indexes
WHERE schemaname = 'hris'
AND indexname = 'idx_employees_nik';
EOF

########################################
# 5. MANIFEST.JSON
########################################
cat > $OUT_DIR/manifest.json << 'EOF'
{
  "bundle_version": "2026.02.20.001",
  "app": "hris-db-migrator-prototype",
  "target_schema_version": "2026.02.20.002",
  "db": {
    "type": "postgres",
    "min_version": 13,
    "default_schema": "hris"
  },
  "baseline": {
    "version": "2026.02.20.000",
    "name": "baseline_initial_hris_schema",
    "file": "baseline/20260220_000_baseline.sql",
    "required_for_fresh_db": true
  },
  "migrations": [
    {
      "version": "2026.02.20.001",
      "name": "add_employee_nik_column",
      "file": "migrations/20260220_001_add_employee_nik.sql",
      "transaction": true
    },
    {
      "version": "2026.02.20.002",
      "name": "create_index_employee_nik",
      "file": "migrations/20260220_002_create_index_employee_nik_notx.sql",
      "transaction": false
    }
  ],
  "checks": {
    "post_migration": [
      "checks/smoke.sql"
    ]
  },
  "execution": {
    "use_advisory_lock": true,
    "lock_key": 987654321,
    "stop_on_error": true
  }
}
EOF

########################################
# 6. GENERATE CHECKSUMS (NO PYTHON)
########################################
echo "==> Generating checksums.json..."

CHECKSUM_FILE="$OUT_DIR/checksums.json"
echo "{" > $CHECKSUM_FILE

FIRST=true
find $OUT_DIR -type f ! -name "checksums.json" | sort | while read file; do
    REL_PATH=${file#"$OUT_DIR/"}

    if command -v sha256sum >/dev/null 2>&1; then
        HASH=$(sha256sum "$file" | awk '{print $1}')
    else
        HASH=$(openssl dgst -sha256 "$file" | awk '{print $2}')
    fi

    if [ "$FIRST" = true ]; then
        FIRST=false
    else
        echo "," >> $CHECKSUM_FILE
    fi

    echo "  \"$REL_PATH\": \"sha256:$HASH\"" >> $CHECKSUM_FILE
done

echo "" >> $CHECKSUM_FILE
echo "}" >> $CHECKSUM_FILE

########################################
# 7. ZIP BUNDLE
########################################
echo "==> Creating zip bundle..."
rm -f $ZIP_NAME
cd $OUT_DIR
zip -r ../$ZIP_NAME . > /dev/null
cd ..

echo "=================================="
echo "Bundle generated successfully!"
echo "Folder : $OUT_DIR/"
echo "Zip    : $ZIP_NAME"
echo "=================================="
