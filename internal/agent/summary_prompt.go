package agent

import (
	"strings"
)

func BuildSummaryPrompt(input SummaryInput) []PromptMessage {
	correctionsJSON := mustMarshalPromptJSON(input.Corrections)
	scoreJSON := mustMarshalPromptJSON(input.Score)

	system := strings.Join([]string{
		"You are SpeakMate's post-practice summary agent.",
		"Generate a structured Chinese after-class report for an English speaking practice session.",
		"Return only valid JSON. Do not wrap it in markdown. Do not add extra text.",
		"Do not invent messages, scores, corrections, or user behavior that is not present in the input.",
		"Frequent errors must come from the correction results or a conservative grouping of those results.",
		"Next practice plan must be concrete, actionable, and related to the speaking scenario.",
		"Return this JSON shape exactly:",
		`{"summary":"...","major_problems":["..."],"frequent_errors":["..."],"better_expressions":["..."],"next_practice_plan":["..."]}`,
	}, "\n")

	user := strings.Join([]string{
		formatScenarioContext(input.Scenario),
		formatSessionContext(input.Session, ""),
		formatHistoryContext(input.HistoryMessages()),
		"Corrections JSON:",
		correctionsJSON,
		"Current score JSON:",
		scoreJSON,
	}, "\n\n")

	return []PromptMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}
