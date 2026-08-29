BEGIN;

DROP INDEX idx_quota_reservations_used_event_offering;

ALTER TABLE offerings
    DROP CONSTRAINT offerings_participant_quota_max_check,
    DROP CONSTRAINT offerings_price_minor_max_check,
    DROP CONSTRAINT offerings_version_max_check,
    DROP CONSTRAINT offerings_version_positive_check,
    DROP COLUMN version;

ALTER TABLE qurban_events
    DROP CONSTRAINT qurban_events_participant_quota_max_check,
    DROP CONSTRAINT qurban_events_version_max_check,
    DROP CONSTRAINT qurban_events_version_positive_check,
    DROP COLUMN version;

COMMIT;
