package agent

import (
	"strings"
	"testing"
)

func TestBuildSummaryPromptRequiresEvidenceBasedDetailedReport(t *testing.T) {
	messages := BuildSummaryPrompt(validSummaryInput())
	joined := joinPromptMessages(messages)

	required := []string{
		"Every report item must cite concrete evidence from the conversation or correction results.",
		"Use the user's exact wording when explaining problems.",
		"Frequent errors must include the original phrase, suggested expression, and reason.",
		"Next practice plan must include a task and a measurable acceptance check.",
	}
	for _, fragment := range required {
		if !strings.Contains(joined, fragment) {
			t.Fatalf("summary prompt missing %q in:\n%s", fragment, joined)
		}
	}
	if !strings.Contains(joined, "I am study computer science") {
		t.Fatalf("summary prompt does not contain conversation history:\n%s", joined)
	}
}
