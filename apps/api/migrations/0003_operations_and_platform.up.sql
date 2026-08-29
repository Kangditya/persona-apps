BEGIN;

CREATE TABLE livestock (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    livestock_code text NOT NULL,
    tag_code text,
    species text NOT NULL,
    category text NOT NULL,
    source_reference text,
    location_id uuid,
    weight_kg numeric(10,3) CHECK (weight_kg IS NULL OR weight_kg > 0),
    status text NOT NULL CHECK (status IN ('REGISTERED', 'INSPECTED', 'READY', 'ALLOCATED', 'QUEUED', 'SLAUGHTERED', 'HELD', 'CANCELLED')),
    inspected_at timestamptz,
    ready_at timestamptz,
    slaughtered_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (event_id, livestock_code),
    FOREIGN KEY (event_id, location_id) REFERENCES event_locations (event_id, id)
);

CREATE TABLE livestock_inspections (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    livestock_id uuid NOT NULL,
    result text NOT NULL CHECK (result IN ('PENDING', 'PASSED', 'FAILED', 'HOLD')),
    notes text,
    inspected_by_operator_id uuid REFERENCES operator_users(id),
    inspected_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    FOREIGN KEY (event_id, livestock_id) REFERENCES livestock (event_id, id)
);

CREATE TABLE livestock_status_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    livestock_id uuid NOT NULL,
    from_status text,
    to_status text NOT NULL CHECK (to_status IN ('REGISTERED', 'INSPECTED', 'READY', 'ALLOCATED', 'QUEUED', 'SLAUGHTERED', 'HELD', 'CANCELLED')),
    reason text,
    changed_by_operator_id uuid REFERENCES operator_users(id),
    changed_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    FOREIGN KEY (event_id, livestock_id) REFERENCES livestock (event_id, id)
);

CREATE TABLE livestock_location_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    livestock_id uuid NOT NULL,
    location_id uuid NOT NULL,
    assigned_by_operator_id uuid REFERENCES operator_users(id),
    assigned_at timestamptz NOT NULL DEFAULT now(),
    released_at timestamptz,
    reason text,
    UNIQUE (event_id, id),
    FOREIGN KEY (event_id, livestock_id) REFERENCES livestock (event_id, id),
    FOREIGN KEY (event_id, location_id) REFERENCES event_locations (event_id, id),
    CHECK (released_at IS NULL OR released_at >= assigned_at)
);

CREATE TABLE allocations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    livestock_id uuid NOT NULL,
    purchase_id uuid,
    sohibul_qurban_id uuid,
    allocation_ref text NOT NULL UNIQUE,
    capacity_units integer NOT NULL DEFAULT 1 CHECK (capacity_units > 0),
    status text NOT NULL CHECK (status IN ('PROVISIONAL', 'CONFIRMED', 'RELEASED', 'REASSIGNED')),
    reason text,
    allocated_at timestamptz NOT NULL DEFAULT now(),
    confirmed_at timestamptz,
    released_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    FOREIGN KEY (event_id, livestock_id) REFERENCES livestock (event_id, id),
    FOREIGN KEY (event_id, purchase_id) REFERENCES purchases (event_id, id),
    FOREIGN KEY (event_id, sohibul_qurban_id) REFERENCES sohibul_qurban (event_id, id),
    CHECK (num_nonnulls(purchase_id, sohibul_qurban_id) >= 1),
    CHECK (released_at IS NULL OR released_at >= allocated_at)
);

CREATE TABLE allocation_status_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    allocation_id uuid NOT NULL REFERENCES allocations(id),
    from_status text,
    to_status text NOT NULL CHECK (to_status IN ('PROVISIONAL', 'CONFIRMED', 'RELEASED', 'REASSIGNED')),
    reason text,
    changed_by_operator_id uuid REFERENCES operator_users(id),
    changed_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE slaughter_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    session_ref text NOT NULL UNIQUE,
    scheduled_start_at timestamptz,
    scheduled_end_at timestamptz,
    status text NOT NULL CHECK (status IN ('PLANNED', 'OPEN', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    CHECK (
        scheduled_start_at IS NULL
        OR scheduled_end_at IS NULL
        OR scheduled_start_at < scheduled_end_at
    )
);

CREATE TABLE slaughter_stations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    session_id uuid NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (session_id, code),
    FOREIGN KEY (event_id, session_id) REFERENCES slaughter_sessions (event_id, id)
);

CREATE TABLE slaughter_records (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    session_id uuid NOT NULL,
    livestock_id uuid NOT NULL,
    allocation_id uuid,
    station_id uuid,
    queue_position integer CHECK (queue_position IS NULL OR queue_position > 0),
    status text NOT NULL CHECK (status IN ('QUEUED', 'CALLED', 'IN_PROGRESS', 'COMPLETED', 'HELD', 'CANCELLED')),
    recorded_by_operator_id uuid REFERENCES operator_users(id),
    checked_in_at timestamptz,
    started_at timestamptz,
    completed_at timestamptz,
    held_reason text,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (event_id, livestock_id),
    FOREIGN KEY (event_id, session_id) REFERENCES slaughter_sessions (event_id, id),
    FOREIGN KEY (event_id, livestock_id) REFERENCES livestock (event_id, id),
    FOREIGN KEY (event_id, allocation_id) REFERENCES allocations (event_id, id),
    FOREIGN KEY (event_id, station_id) REFERENCES slaughter_stations (event_id, id),
    CHECK (completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at)
);

CREATE TABLE distribution_records (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    purchase_id uuid,
    sohibul_qurban_id uuid,
    distribution_ref text NOT NULL UNIQUE,
    method text NOT NULL,
    status text NOT NULL CHECK (status IN ('PENDING', 'PREPARED', 'READY', 'COLLECTED', 'DELIVERED', 'FAILED', 'CANCELLED')),
    prepared_at timestamptz,
    completed_at timestamptz,
    failure_reason text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    FOREIGN KEY (event_id, purchase_id) REFERENCES purchases (event_id, id),
    FOREIGN KEY (event_id, sohibul_qurban_id) REFERENCES sohibul_qurban (event_id, id),
    CHECK (num_nonnulls(purchase_id, sohibul_qurban_id) >= 1),
    CHECK (completed_at IS NULL OR prepared_at IS NULL OR completed_at >= prepared_at)
);

CREATE TABLE distribution_status_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    distribution_id uuid NOT NULL REFERENCES distribution_records(id),
    from_status text,
    to_status text NOT NULL CHECK (to_status IN ('PENDING', 'PREPARED', 'READY', 'COLLECTED', 'DELIVERED', 'FAILED', 'CANCELLED')),
    reason text,
    changed_by_operator_id uuid REFERENCES operator_users(id),
    changed_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_operator_id uuid REFERENCES operator_users(id),
    actor_party_id uuid REFERENCES parties(id),
    actor_reference text NOT NULL,
    action text NOT NULL,
    target_type text NOT NULL,
    target_id uuid,
    request_id text,
    reason text,
    before_data jsonb,
    after_data jsonb,
    source_application text,
    client_ip inet,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (num_nonnulls(actor_operator_id, actor_party_id) <= 1)
);

CREATE TABLE outbox_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type text NOT NULL,
    aggregate_id uuid NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    occurred_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE idempotency_records (
    namespace text NOT NULL,
    idempotency_key text NOT NULL,
    request_hash text NOT NULL,
    response_status integer,
    response_body jsonb,
    expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (namespace, idempotency_key)
);

CREATE UNIQUE INDEX uq_livestock_tag_code
    ON livestock (event_id, tag_code)
    WHERE tag_code IS NOT NULL;

CREATE UNIQUE INDEX uq_livestock_current_location
    ON livestock_location_history (livestock_id)
    WHERE released_at IS NULL;

CREATE UNIQUE INDEX uq_active_participant_allocation
    ON allocations (sohibul_qurban_id)
    WHERE sohibul_qurban_id IS NOT NULL AND status IN ('PROVISIONAL', 'CONFIRMED');

CREATE INDEX idx_livestock_status ON livestock (event_id, status);
CREATE INDEX idx_livestock_location ON livestock (event_id, location_id) WHERE location_id IS NOT NULL;
CREATE INDEX idx_livestock_inspections ON livestock_inspections (livestock_id, inspected_at);
CREATE INDEX idx_livestock_status_history ON livestock_status_history (livestock_id, changed_at);
CREATE INDEX idx_livestock_location_history ON livestock_location_history (livestock_id, assigned_at);
CREATE INDEX idx_allocations_status ON allocations (event_id, status);
CREATE INDEX idx_allocations_livestock ON allocations (livestock_id, status);
CREATE INDEX idx_allocations_purchase ON allocations (purchase_id, status) WHERE purchase_id IS NOT NULL;
CREATE INDEX idx_allocations_participant ON allocations (sohibul_qurban_id, status) WHERE sohibul_qurban_id IS NOT NULL;
CREATE INDEX idx_allocation_status_history ON allocation_status_history (allocation_id, changed_at);
CREATE INDEX idx_slaughter_sessions_schedule ON slaughter_sessions (event_id, status, scheduled_start_at);
CREATE INDEX idx_slaughter_stations_session ON slaughter_stations (session_id, code);
CREATE INDEX idx_slaughter_records_queue ON slaughter_records (session_id, status, queue_position);
CREATE INDEX idx_slaughter_records_livestock ON slaughter_records (livestock_id);
CREATE INDEX idx_distribution_status ON distribution_records (event_id, status);
CREATE INDEX idx_distribution_purchase ON distribution_records (purchase_id, status) WHERE purchase_id IS NOT NULL;
CREATE INDEX idx_distribution_participant ON distribution_records (sohibul_qurban_id, status) WHERE sohibul_qurban_id IS NOT NULL;
CREATE INDEX idx_distribution_status_history ON distribution_status_history (distribution_id, changed_at);
CREATE INDEX idx_audit_target ON audit_log (target_type, target_id, created_at);
CREATE INDEX idx_audit_actor ON audit_log (actor_reference, created_at);
CREATE INDEX idx_audit_request ON audit_log (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX idx_outbox_unpublished ON outbox_events (created_at, id) WHERE published_at IS NULL;
CREATE INDEX idx_outbox_aggregate ON outbox_events (aggregate_type, aggregate_id, created_at);
CREATE INDEX idx_idempotency_expiry ON idempotency_records (expires_at) WHERE expires_at IS NOT NULL;

COMMIT;
