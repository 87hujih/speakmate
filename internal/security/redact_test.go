package security

import (
	"strings"
	"testing"
)

func TestRedactStringHidesSensitiveKeyValuesAndBearerTokens(t *testing.T) {
	input := "LLM_API_KEY=test-key password=plain-secret Authorization: Bearer token-value"

	got := RedactString(input)

	for _, secret := range []string{"test-key", "plain-secret", "token-value"} {
		if strings.Contains(got, secret) {
			t.Fatalf("RedactString output %q still contains secret %q", got, secret)
		}
	}
	if strings.Count(got, "[REDACTED]") < 3 {
		t.Fatalf("RedactString output = %q, want redaction markers", got)
	}
}

func TestRedactStringHidesMySQLDSNPassword(t *testing.T) {
	input := "speakmate:db-secret@tcp(mysql:3306)/speakmate?parseTime=true"

	got := RedactString(input)

	if strings.Contains(got, "db-secret") {
		t.Fatalf("RedactString output %q still contains DSN password", got)
	}
	if !strings.Contains(got, "speakmate:[REDACTED]@tcp(mysql:3306)") {
		t.Fatalf("RedactString output = %q, want DSN password redacted", got)
	}
}

func TestRedactURLHidesSensitiveQueryValues(t *testing.T) {
	input := "/api/v1/scenarios?api_key=test-key&token=test-token&password=test-password&q=visible"

	got := RedactURL(input)

	for _, secret := range []string{"test-key", "test-token", "test-password"} {
		if strings.Contains(got, secret) {
			t.Fatalf("RedactURL output %q still contains secret %q", got, secret)
		}
	}
	if !strings.Contains(got, "q=visible") {
		t.Fatalf("RedactURL output = %q, want non-sensitive query value preserved", got)
	}
}
