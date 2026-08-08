BEGIN;

DROP TABLE idempotency_records;
DROP TABLE outbox_events;
DROP TABLE audit_log;
DROP TABLE distribution_status_history;
DROP TABLE distribution_records;
DROP TABLE slaughter_records;
DROP TABLE slaughter_stations;
DROP TABLE slaughter_sessions;
DROP TABLE allocation_status_history;
DROP TABLE allocations;
DROP TABLE livestock_location_history;
DROP TABLE livestock_status_history;
DROP TABLE livestock_inspections;
DROP TABLE livestock;

COMMIT;
