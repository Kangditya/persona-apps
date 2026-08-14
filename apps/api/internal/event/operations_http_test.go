package event

import (
    "encoding/json"
    "errors"
    "testing"
)

func TestPatchEventRequestPreservesOmittedAndNullFields(t *testing.T) {
    var request patchEventRequest
    if err := json.Unmarshal([]byte(`{"expected_version":4,"registration_opens_at":null,"participant_quota":null}`), &request); err != nil {
        t.Fatal(err)
    }
    input, err := request.input()
    if err != nil {
        t.Fatal(err)
    }
    if input.ExpectedVersion != 4 || input.Name != nil || !input.RegistrationOpensAt.Set || input.RegistrationOpensAt.Value != nil || input.RegistrationClosesAt.Set || !input.ParticipantQuota.Set || input.ParticipantQuota.Value != nil {
        t.Fatalf("patch input = %#v", input)
    }

    request = patchEventRequest{}
    if err := json.Unmarshal([]byte(`{"expected_version":4}`), &request); err != nil {
        t.Fatal(err)
    }
    if _, err := request.input(); !errors.Is(err, ErrNoChanges) {
        t.Fatalf("expected-version-only error = %v", err)
    }
}
