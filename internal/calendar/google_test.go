package calendar

import "testing"

func TestClientOptionsValidateServiceAccount(t *testing.T) {
	opts := ClientOptions{
		CalendarID:            "calendar@example.com",
		ServiceAccountKeyFile: "/run/secrets/google-sa.json",
	}
	if err := opts.Validate(); err != nil {
		t.Fatalf("service account key file should be valid: %v", err)
	}

	opts = ClientOptions{
		CalendarID:            "calendar@example.com",
		ServiceAccountKeyJSON: `{"client_email":"sa@example.com"}`,
	}
	if err := opts.Validate(); err != nil {
		t.Fatalf("service account key json should be valid: %v", err)
	}
}

func TestClientOptionsValidateRequiresCredentials(t *testing.T) {
	opts := ClientOptions{CalendarID: "calendar@example.com"}
	if err := opts.Validate(); err == nil {
		t.Fatalf("missing service account credentials should be invalid")
	}
}

func TestClientOptionsFingerprintIncludesServiceAccount(t *testing.T) {
	base := ClientOptions{
		CalendarID:            "calendar@example.com",
		ServiceAccountKeyFile: "/run/secrets/google-sa.json",
	}
	changed := ClientOptions{
		CalendarID:            "calendar@example.com",
		ServiceAccountKeyFile: "/run/secrets/google-sa-2.json",
	}
	if base.Fingerprint() == changed.Fingerprint() {
		t.Fatalf("fingerprint must change when service account key file changes")
	}
}
