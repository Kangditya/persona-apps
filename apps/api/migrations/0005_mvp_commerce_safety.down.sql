BEGIN;

ALTER TABLE audit_log
    DROP COLUMN permission,
    DROP COLUMN source;

DROP TABLE operator_sessions;

ALTER TABLE payment_records
    DROP CONSTRAINT payment_records_evidence_metadata_check,
    DROP COLUMN evidence_sha256,
    DROP COLUMN evidence_size_bytes,
    DROP COLUMN evidence_media_type,
    DROP COLUMN evidence_filename;

DROP TABLE quota_reservations;
DROP TABLE purchase_participants;

DROP INDEX uq_purchases_access_token_hash;

ALTER TABLE purchases
    DROP CONSTRAINT purchases_common_access_token_check,
    DROP CONSTRAINT purchases_access_token_hash_sha256_check,
    DROP CONSTRAINT uq_purchases_event_id_offering,
    DROP COLUMN access_token_hash;

ALTER TABLE offerings
    DROP CONSTRAINT offerings_participant_quota_check,
    DROP COLUMN participant_quota;

DROP INDEX uq_qurban_events_one_active;

ALTER TABLE qurban_events
    DROP CONSTRAINT qurban_events_status_check,
    ADD CONSTRAINT qurban_events_status_check
        CHECK (status IN ('DRAFT', 'PUBLISHED', 'ACTIVE', 'CLOSED', 'ARCHIVED'));

COMMIT;
