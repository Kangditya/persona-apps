BEGIN;

CREATE TABLE schema_seeds (
    name text PRIMARY KEY,
    executed_at timestamptz NOT NULL DEFAULT now()
);

COMMIT;
