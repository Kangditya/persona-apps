BEGIN;

CREATE TABLE saving_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    account_ref text NOT NULL UNIQUE,
    holder_party_id uuid NOT NULL REFERENCES parties(id),
    target_offering_id uuid,
    target_amount_minor bigint NOT NULL CHECK (target_amount_minor > 0),
    currency_code text NOT NULL CHECK (currency_code ~ '^[A-Z]{3}$'),
    status text NOT NULL CHECK (status IN ('DRAFT', 'ACTIVE', 'PARTIALLY_FUNDED', 'FULLY_FUNDED', 'CONVERSION_PENDING', 'CONVERTED', 'CANCELLED', 'EXPIRED')),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    fully_funded_at timestamptz,
    converted_at timestamptz,
    cancelled_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    FOREIGN KEY (event_id, target_offering_id) REFERENCES offerings (event_id, id)
);

CREATE TABLE giveaway_programs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    code text NOT NULL,
    sponsor_party_id uuid NOT NULL REFERENCES parties(id),
    name text NOT NULL,
    description text,
    funding_limit_minor bigint CHECK (funding_limit_minor IS NULL OR funding_limit_minor >= 0),
    currency_code text NOT NULL CHECK (currency_code ~ '^[A-Z]{3}$'),
    status text NOT NULL CHECK (status IN ('DRAFT', 'OPEN', 'CLOSED', 'COMPLETED', 'CANCELLED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (event_id, code)
);

CREATE TABLE giveaway_applications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id uuid NOT NULL REFERENCES giveaway_programs(id),
    application_ref text NOT NULL UNIQUE,
    applicant_party_id uuid REFERENCES parties(id),
    nominee_party_id uuid REFERENCES parties(id),
    status text NOT NULL CHECK (status IN ('SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED', 'ASSIGNED', 'WITHDRAWN')),
    reviewed_by_operator_id uuid REFERENCES operator_users(id),
    reviewed_at timestamptz,
    review_notes text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (program_id, id),
    CHECK (num_nonnulls(applicant_party_id, nominee_party_id) = 1)
);

CREATE TABLE giveaway_assignments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    program_id uuid NOT NULL,
    application_id uuid NOT NULL,
    recipient_party_id uuid NOT NULL REFERENCES parties(id),
    status text NOT NULL CHECK (status IN ('APPROVED', 'ASSIGNED', 'CANCELLED')),
    assigned_by_operator_id uuid REFERENCES operator_users(id),
    assigned_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (program_id, application_id),
    UNIQUE (program_id, recipient_party_id),
    FOREIGN KEY (event_id, program_id) REFERENCES giveaway_programs (event_id, id),
    FOREIGN KEY (program_id, application_id) REFERENCES giveaway_applications (program_id, id)
);

CREATE TABLE purchases (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    purchase_ref text NOT NULL UNIQUE,
    channel text NOT NULL CHECK (channel IN ('COMMON', 'SAVING', 'GIVEAWAY')),
    purchaser_party_id uuid NOT NULL REFERENCES parties(id),
    payer_party_id uuid REFERENCES parties(id),
    offering_id uuid NOT NULL,
    saving_account_id uuid,
    giveaway_assignment_id uuid,
    participant_count integer NOT NULL CHECK (participant_count > 0),
    offering_name_snapshot text NOT NULL,
    offering_kind_snapshot text NOT NULL,
    offering_unit_price_minor bigint NOT NULL CHECK (offering_unit_price_minor >= 0),
    participant_capacity_snapshot integer NOT NULL CHECK (participant_capacity_snapshot > 0),
    total_amount_minor bigint NOT NULL CHECK (total_amount_minor >= 0),
    currency_code text NOT NULL CHECK (currency_code ~ '^[A-Z]{3}$'),
    status text NOT NULL CHECK (status IN ('DRAFT', 'PENDING_PAYMENT', 'PAID', 'ELIGIBLE', 'ALLOCATED', 'COMPLETED', 'CANCELLED')),
    eligible_at timestamptz,
    cancelled_at timestamptz,
    cancellation_reason text,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    FOREIGN KEY (event_id, offering_id) REFERENCES offerings (event_id, id),
    FOREIGN KEY (event_id, saving_account_id) REFERENCES saving_accounts (event_id, id),
    FOREIGN KEY (event_id, giveaway_assignment_id) REFERENCES giveaway_assignments (event_id, id),
    CHECK (
        (channel = 'COMMON' AND saving_account_id IS NULL AND giveaway_assignment_id IS NULL)
        OR (channel = 'SAVING' AND saving_account_id IS NOT NULL AND giveaway_assignment_id IS NULL)
        OR (channel = 'GIVEAWAY' AND saving_account_id IS NULL AND giveaway_assignment_id IS NOT NULL)
    )
);

CREATE TABLE purchase_status_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_id uuid NOT NULL REFERENCES purchases(id),
    from_status text,
    to_status text NOT NULL CHECK (to_status IN ('DRAFT', 'PENDING_PAYMENT', 'PAID', 'ELIGIBLE', 'ALLOCATED', 'COMPLETED', 'CANCELLED')),
    reason text,
    changed_by_operator_id uuid REFERENCES operator_users(id),
    changed_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sohibul_qurban (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    purchase_id uuid NOT NULL,
    party_id uuid NOT NULL REFERENCES parties(id),
    participant_ref text NOT NULL UNIQUE,
    display_name_snapshot text NOT NULL,
    sequence_no integer NOT NULL CHECK (sequence_no > 0),
    status text NOT NULL CHECK (status IN ('PENDING', 'ACTIVE', 'REPLACED', 'CANCELLED')),
    verified_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    UNIQUE (purchase_id, sequence_no),
    FOREIGN KEY (event_id, purchase_id) REFERENCES purchases (event_id, id)
);

CREATE TABLE sohibul_qurban_status_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    sohibul_qurban_id uuid NOT NULL REFERENCES sohibul_qurban(id),
    from_status text,
    to_status text NOT NULL CHECK (to_status IN ('PENDING', 'ACTIVE', 'REPLACED', 'CANCELLED')),
    reason text,
    changed_by_operator_id uuid REFERENCES operator_users(id),
    changed_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE payment_records (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    payment_ref text NOT NULL UNIQUE,
    payer_party_id uuid NOT NULL REFERENCES parties(id),
    purchase_id uuid,
    saving_account_id uuid,
    giveaway_program_id uuid,
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency_code text NOT NULL CHECK (currency_code ~ '^[A-Z]{3}$'),
    method text NOT NULL,
    status text NOT NULL CHECK (status IN ('SUBMITTED', 'VERIFIED', 'REJECTED', 'REFUNDED')),
    provider_reference text,
    evidence_reference text,
    submitted_at timestamptz NOT NULL DEFAULT now(),
    verified_at timestamptz,
    verified_by_operator_id uuid REFERENCES operator_users(id),
    rejection_reason text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    FOREIGN KEY (event_id, purchase_id) REFERENCES purchases (event_id, id),
    FOREIGN KEY (event_id, saving_account_id) REFERENCES saving_accounts (event_id, id),
    FOREIGN KEY (event_id, giveaway_program_id) REFERENCES giveaway_programs (event_id, id),
    CHECK (num_nonnulls(purchase_id, saving_account_id, giveaway_program_id) = 1)
);

CREATE TABLE payment_status_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id uuid NOT NULL REFERENCES payment_records(id),
    from_status text,
    to_status text NOT NULL CHECK (to_status IN ('SUBMITTED', 'VERIFIED', 'REJECTED', 'REFUNDED')),
    reason text,
    changed_by_operator_id uuid REFERENCES operator_users(id),
    changed_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE financial_ledger_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES qurban_events(id),
    entry_type text NOT NULL CHECK (entry_type IN ('PAYMENT', 'INSTALLMENT', 'SPONSOR_FUNDING', 'REFUND', 'TRANSFER', 'ADJUSTMENT')),
    direction text NOT NULL CHECK (direction IN ('CREDIT', 'DEBIT')),
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency_code text NOT NULL CHECK (currency_code ~ '^[A-Z]{3}$'),
    purchase_id uuid,
    saving_account_id uuid,
    giveaway_program_id uuid,
    payment_id uuid,
    idempotency_key text,
    correlation_ref text,
    recorded_by_operator_id uuid REFERENCES operator_users(id),
    reason text,
    occurred_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (event_id, id),
    FOREIGN KEY (event_id, purchase_id) REFERENCES purchases (event_id, id),
    FOREIGN KEY (event_id, saving_account_id) REFERENCES saving_accounts (event_id, id),
    FOREIGN KEY (event_id, giveaway_program_id) REFERENCES giveaway_programs (event_id, id),
    FOREIGN KEY (event_id, payment_id) REFERENCES payment_records (event_id, id),
    CHECK (num_nonnulls(purchase_id, saving_account_id, giveaway_program_id) = 1)
);

CREATE UNIQUE INDEX uq_payment_provider_reference
    ON payment_records (provider_reference)
    WHERE provider_reference IS NOT NULL;

CREATE UNIQUE INDEX uq_financial_ledger_idempotency_key
    ON financial_ledger_entries (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE UNIQUE INDEX uq_purchase_saving_account
    ON purchases (saving_account_id)
    WHERE saving_account_id IS NOT NULL;

CREATE UNIQUE INDEX uq_purchase_giveaway_assignment
    ON purchases (giveaway_assignment_id)
    WHERE giveaway_assignment_id IS NOT NULL;

CREATE INDEX idx_saving_accounts_status ON saving_accounts (event_id, status);
CREATE INDEX idx_saving_accounts_holder ON saving_accounts (holder_party_id, status);
CREATE INDEX idx_giveaway_programs_status ON giveaway_programs (event_id, status);
CREATE INDEX idx_giveaway_applications_status ON giveaway_applications (program_id, status);
CREATE INDEX idx_giveaway_assignments_status ON giveaway_assignments (event_id, status);
CREATE INDEX idx_purchases_status ON purchases (event_id, status);
CREATE INDEX idx_purchases_channel ON purchases (event_id, channel);
CREATE INDEX idx_purchases_purchaser ON purchases (purchaser_party_id, created_at);
CREATE INDEX idx_purchases_payer ON purchases (payer_party_id, created_at) WHERE payer_party_id IS NOT NULL;
CREATE INDEX idx_purchase_status_history_purchase ON purchase_status_history (purchase_id, changed_at);
CREATE INDEX idx_sohibul_qurban_status ON sohibul_qurban (event_id, status);
CREATE INDEX idx_sohibul_qurban_purchase ON sohibul_qurban (purchase_id, sequence_no);
CREATE INDEX idx_sohibul_qurban_status_history ON sohibul_qurban_status_history (sohibul_qurban_id, changed_at);
CREATE INDEX idx_payment_records_status ON payment_records (event_id, status, submitted_at);
CREATE INDEX idx_payment_records_purchase ON payment_records (purchase_id, submitted_at) WHERE purchase_id IS NOT NULL;
CREATE INDEX idx_payment_records_saving ON payment_records (saving_account_id, submitted_at) WHERE saving_account_id IS NOT NULL;
CREATE INDEX idx_payment_records_program ON payment_records (giveaway_program_id, submitted_at) WHERE giveaway_program_id IS NOT NULL;
CREATE INDEX idx_payment_status_history_payment ON payment_status_history (payment_id, changed_at);
CREATE INDEX idx_financial_ledger_event_time ON financial_ledger_entries (event_id, occurred_at);
CREATE INDEX idx_financial_ledger_purchase ON financial_ledger_entries (purchase_id, occurred_at) WHERE purchase_id IS NOT NULL;
CREATE INDEX idx_financial_ledger_saving ON financial_ledger_entries (saving_account_id, occurred_at) WHERE saving_account_id IS NOT NULL;
CREATE INDEX idx_financial_ledger_program ON financial_ledger_entries (giveaway_program_id, occurred_at) WHERE giveaway_program_id IS NOT NULL;

COMMIT;
