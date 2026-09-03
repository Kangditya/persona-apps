BEGIN;

CREATE UNIQUE INDEX uq_payment_records_one_submitted_purchase
    ON payment_records (purchase_id)
    WHERE purchase_id IS NOT NULL AND status = 'SUBMITTED';

COMMIT;
