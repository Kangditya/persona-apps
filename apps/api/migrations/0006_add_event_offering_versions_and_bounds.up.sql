BEGIN;

ALTER TABLE qurban_events
    ADD COLUMN version bigint NOT NULL DEFAULT 1,
    ADD CONSTRAINT qurban_events_version_positive_check CHECK (version > 0),
    ADD CONSTRAINT qurban_events_version_max_check CHECK (version <= 9007199254740991),
    ADD CONSTRAINT qurban_events_participant_quota_max_check
        CHECK (participant_quota IS NULL OR participant_quota <= 9007199254740991);

ALTER TABLE offerings
    ADD COLUMN version bigint NOT NULL DEFAULT 1,
    ADD CONSTRAINT offerings_version_positive_check CHECK (version > 0),
    ADD CONSTRAINT offerings_version_max_check CHECK (version <= 9007199254740991),
    ADD CONSTRAINT offerings_price_minor_max_check CHECK (price_minor <= 9007199254740991),
    ADD CONSTRAINT offerings_participant_quota_max_check
        CHECK (participant_quota IS NULL OR participant_quota <= 9007199254740991);

CREATE INDEX idx_quota_reservations_used_event_offering
    ON quota_reservations (event_id, offering_id)
    WHERE status IN ('RESERVED', 'CONSUMED');

COMMIT;
