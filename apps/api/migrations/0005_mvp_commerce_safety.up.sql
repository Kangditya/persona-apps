BEGIN;

ALTER TABLE qurban_events
    DROP CONSTRAINT qurban_events_status_check,
    ADD CONSTRAINT qurban_events_status_check
        CHECK (status IN ('DRAFT', 'PUBLISHED', 'ACTIVE', 'SUSPENDED', 'CLOSED', 'ARCHIVED'));

CREATE UNIQUE INDEX uq_qurban_events_one_active
    ON qurban_events (status)
    WHERE status = 'ACTIVE';

ALTER TABLE offerings
    ADD COLUMN participant_quota bigint,
    ADD CONSTRAINT offerings_participant_quota_check
        CHECK (participant_quota IS NULL OR participant_quota >= 0);

ALTER TABLE purchases
    ADD COLUMN access_token_hash bytea,
    ADD CONSTRAINT purchases_access_token_hash_sha256_check
        CHECK (access_token_hash IS NULL OR octet_length(access_token_hash) = 32),
    ADD CONSTRAINT purchases_common_access_token_check
        CHECK (channel <> 'COMMON' OR access_token_hash IS NOT NULL),
    ADD CONSTRAINT uq_purchases_event_id_offering
        UNIQUE (event_id, id, offering_id);

CREATE UNIQUE INDEX uq_purchases_access_token_hash
    ON purchases (access_token_hash)
    WHERE access_token_hash IS NOT NULL;

CREATE TABLE purchase_participants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    purchase_id uuid NOT NULL,
    party_id uuid REFERENCES parties(id),
    sequence_no integer NOT NULL CHECK (sequence_no > 0),
    display_name_snapshot text NOT NULL CHECK (btrim(display_name_snapshot) <> ''),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (purchase_id, sequence_no),
    FOREIGN KEY (event_id, purchase_id) REFERENCES purchases (event_id, id)
);

CREATE TABLE quota_reservations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    purchase_id uuid NOT NULL,
    offering_id uuid NOT NULL,
    attempt_no integer NOT NULL CHECK (attempt_no > 0),
    participant_units integer NOT NULL CHECK (participant_units > 0),
    status text NOT NULL CHECK (status IN ('RESERVED', 'CONSUMED', 'RELEASED', 'EXPIRED')),
    expires_at timestamptz,
    consumed_at timestamptz,
    released_at timestamptz,
    release_reason text,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (purchase_id, attempt_no),
    FOREIGN KEY (event_id, purchase_id, offering_id)
        REFERENCES purchases (event_id, id, offering_id),
    CHECK (
        (status = 'RESERVED' AND consumed_at IS NULL AND released_at IS NULL AND release_reason IS NULL)
        OR (status = 'CONSUMED' AND consumed_at IS NOT NULL AND released_at IS NULL AND release_reason IS NULL)
        OR (status IN ('RELEASED', 'EXPIRED') AND consumed_at IS NULL AND released_at IS NOT NULL)
    ),
    CHECK (status <> 'EXPIRED' OR expires_at IS NOT NULL),
    CHECK (consumed_at IS NULL OR consumed_at >= created_at),
    CHECK (released_at IS NULL OR released_at >= created_at)
);

CREATE UNIQUE INDEX uq_quota_reservations_purchase_reserved
    ON quota_reservations (purchase_id)
    WHERE status = 'RESERVED';

CREATE INDEX idx_quota_reservations_active
    ON quota_reservations (event_id, offering_id)
    WHERE status = 'RESERVED';

CREATE INDEX idx_quota_reservations_expiring
    ON quota_reservations (expires_at)
    WHERE status = 'RESERVED' AND expires_at IS NOT NULL;

ALTER TABLE payment_records
    ADD COLUMN evidence_filename text,
    ADD COLUMN evidence_media_type text,
    ADD COLUMN evidence_size_bytes bigint,
    ADD COLUMN evidence_sha256 bytea,
    ADD CONSTRAINT payment_records_evidence_metadata_check CHECK (
        (
            evidence_reference IS NULL
            AND evidence_filename IS NULL
            AND evidence_media_type IS NULL
            AND evidence_size_bytes IS NULL
            AND evidence_sha256 IS NULL
        )
        OR (
            evidence_reference IS NOT NULL
            AND evidence_reference <> ''
            AND evidence_filename IS NOT NULL
            AND btrim(evidence_filename) <> ''
            AND evidence_media_type IN ('image/jpeg', 'image/png', 'application/pdf')
            AND evidence_size_bytes BETWEEN 1 AND 10485760
            AND evidence_sha256 IS NOT NULL
            AND octet_length(evidence_sha256) = 32
        )
    );

CREATE TABLE operator_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_user_id uuid NOT NULL REFERENCES operator_users(id),
    session_token_hash bytea NOT NULL UNIQUE CHECK (octet_length(session_token_hash) = 32),
    csrf_token_hash bytea NOT NULL CHECK (octet_length(csrf_token_hash) = 32),
    permission_snapshot jsonb NOT NULL CHECK (jsonb_typeof(permission_snapshot) = 'array'),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at),
    CHECK (revoked_at IS NULL OR revoked_at >= created_at),
    CHECK (last_seen_at >= created_at)
);

CREATE INDEX idx_operator_sessions_active
    ON operator_sessions (operator_user_id, expires_at)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_operator_sessions_expiry
    ON operator_sessions (expires_at)
    WHERE revoked_at IS NULL;

ALTER TABLE audit_log
    ADD COLUMN source text,
    ADD COLUMN permission text;

COMMIT;
