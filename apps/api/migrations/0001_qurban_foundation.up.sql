BEGIN;

CREATE TABLE operator_users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    external_subject text NOT NULL UNIQUE,
    display_name text NOT NULL,
    status text NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE parties (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    party_type text NOT NULL CHECK (party_type IN ('PERSON', 'ORGANIZATION')),
    display_name text NOT NULL,
    email text,
    phone text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE qurban_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_year smallint NOT NULL UNIQUE CHECK (event_year BETWEEN 1900 AND 9999),
    name text NOT NULL,
    status text NOT NULL CHECK (status IN ('DRAFT', 'PUBLISHED', 'ACTIVE', 'CLOSED', 'ARCHIVED')),
    registration_opens_at timestamptz,
    registration_closes_at timestamptz,
    participant_quota bigint CHECK (participant_quota IS NULL OR participant_quota >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        registration_opens_at IS NULL
        OR registration_closes_at IS NULL
        OR registration_opens_at < registration_closes_at
    )
);

CREATE TABLE event_locations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    code text NOT NULL,
    name text NOT NULL,
    location_type text NOT NULL CHECK (location_type IN ('PEN', 'STATION', 'DISTRIBUTION', 'OTHER')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (event_id, code)
);

CREATE TABLE offerings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    code text NOT NULL,
    name text NOT NULL,
    offering_kind text NOT NULL,
    description text,
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    currency_code text NOT NULL CHECK (currency_code ~ '^[A-Z]{3}$'),
    participant_capacity integer NOT NULL CHECK (participant_capacity > 0),
    status text NOT NULL CHECK (status IN ('DRAFT', 'PUBLISHED', 'UNAVAILABLE', 'ARCHIVED')),
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (event_id, code)
);

CREATE INDEX idx_parties_email ON parties (email) WHERE email IS NOT NULL;
CREATE INDEX idx_parties_phone ON parties (phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_qurban_events_status ON qurban_events (status, event_year);
CREATE INDEX idx_event_locations_type ON event_locations (event_id, location_type);
CREATE INDEX idx_offerings_status ON offerings (event_id, status);

COMMIT;
