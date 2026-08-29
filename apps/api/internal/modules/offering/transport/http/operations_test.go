package http

import (
    "encoding/json"
    "errors"
    "testing"
)

func TestPatchOfferingRequestPreservesOmittedAndNullFields(t *testing.T) {
    var request patchOfferingRequest
    if err := json.Unmarshal([]byte(`{"expected_version":7,"description":null,"participant_quota":null}`), &request); err != nil {
        t.Fatal(err)
    }
    input, err := request.input()
    if err != nil {
        t.Fatal(err)
    }
    if input.ExpectedVersion != 7 || input.Name != nil || !input.Description.Set || input.Description.Value != nil || input.PriceMinor != nil || !input.ParticipantQuota.Set || input.ParticipantQuota.Value != nil {
        t.Fatalf("patch input = %#v", input)
    }

    request = patchOfferingRequest{}
    if err := json.Unmarshal([]byte(`{"expected_version":7}`), &request); err != nil {
        t.Fatal(err)
    }
    if _, err := request.input(); !errors.Is(err, ErrNoChanges) {
        t.Fatalf("expected-version-only error = %v", err)
    }
}
