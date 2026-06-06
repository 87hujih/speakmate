package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"speakmate/internal/model"
)

func BuildCorrectionPrompt(input CorrectionInput) []PromptMessage {
	system := strings.Join([]string{
		"You are SpeakMate's English expression correction agent.",
		"Analyze only the user's latest English message in the given speaking scenario.",
		"Return only valid JSON. Do not wrap it in markdown. Do not add extra text.",
		"Use these error types only: grammar, vocabulary, expression, structure, scenario.",
		"Every error must include type, span, suggestion, and explanation.",
		"Explanations should be concise Chinese explanations for the learner.",
		"Return this JSON shape exactly:",
		`{"message_id":1001,"original_text":"...","corrected_text":"...","errors":[{"type":"grammar","span":"...","suggestion":"...","explanation":"..."}],"better_expressions":["..."]}`,
	}, "\n")

	user := strings.Join([]string{
		formatScenarioContext(input.Scenario),
		formatSessionContext(input.Session, input.UserMessage.Stage),
		formatHistoryContext(input.HistoryMessages()),
		fmt.Sprintf("message_id: %d", input.UserMessage.ID),
		"User's latest message:",
		strings.TrimSpace(input.UserMessage.Content),
	}, "\n\n")

	return []PromptMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}

func BuildScoringPrompt(input ScoringInput) []PromptMessage {
	correctionJSON := mustMarshalPromptJSON(input.Correction)
	system := strings.Join([]string{
		"You are SpeakMate's English speaking scoring agent.",
		"Score the user's latest message for the current speaking scenario.",
		"Return only valid JSON. Do not wrap it in markdown. Do not add extra text.",
		"All score fields must be integers from 0 to 100.",
		"Calculate total_score using this weighting: 0.25 fluency, 0.25 grammar, 0.20 expression, 0.15 vocabulary, 0.15 completion.",
		"Return this JSON shape exactly:",
		`{"message_id":1001,"fluency":75,"grammar":72,"expression":80,"vocabulary":76,"completion":85,"total_score":77,"comment":"..."}`,
	}, "\n")

	user := strings.Join([]string{
		formatScenarioContext(input.Scenario),
		formatSessionContext(input.Session, input.UserMessage.Stage),
		formatRubricContext(input.Scenario.Rubric),
		formatHistoryContext(input.HistoryMessages()),
		fmt.Sprintf("message_id: %d", input.UserMessage.ID),
		"User's latest message:",
		strings.TrimSpace(input.UserMessage.Content),
		"Correction result JSON:",
		correctionJSON,
	}, "\n\n")

	return []PromptMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}

func (input CorrectionInput) HistoryMessages() []model.Message {
	if input.History != nil {
		return input.History
	}

	return input.Session.Messages
}

func (input ScoringInput) HistoryMessages() []model.Message {
	if input.History != nil {
		return input.History
	}

	return input.Session.Messages
}

func formatScenarioContext(scenario model.Scenario) string {
	lines := []string{
		"Scenario:",
		"code: " + scenario.Code,
		"name: " + scenario.Name,
		"description: " + scenario.Description,
		"ai_role: " + scenario.AIRole,
		"user_goal: " + scenario.UserGoal,
	}

	if len(scenario.Stages) > 0 {
		lines = append(lines, "stages:")
		for _, stage := range scenario.Stages {
			lines = append(lines, fmt.Sprintf("- %s: %s", stage.Name, stage.Description))
		}
	}

	return strings.Join(nonEmptyLines(lines), "\n")
}

func formatSessionContext(session model.Session, currentStage string) string {
	lines := []string{
		fmt.Sprintf("session_id: %d", session.ID),
		fmt.Sprintf("turn_count: %d", session.TurnCount),
		"current_stage: " + currentStage,
	}

	return strings.Join(nonEmptyLines(lines), "\n")
}

func formatRubricContext(rubric []model.ScenarioRubric) string {
	if len(rubric) == 0 {
		return "Rubric:\n- Use the standard SpeakMate scoring dimensions."
	}

	lines := []string{"Rubric:"}
	for _, item := range rubric {
		lines = append(lines, fmt.Sprintf("- %s: %s", item.Name, item.Description))
	}

	return strings.Join(nonEmptyLines(lines), "\n")
}

func formatHistoryContext(history []model.Message) string {
	if len(history) == 0 {
		return "Recent conversation history: none"
	}

	lines := []string{"Recent conversation history:"}
	for _, message := range history {
		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}
		stage := strings.TrimSpace(message.Stage)
		if stage != "" {
			content = fmt.Sprintf("[%s] %s", stage, content)
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", promptRole(message.Role), content))
	}
	if len(lines) == 1 {
		return "Recent conversation history: none"
	}

	return strings.Join(lines, "\n")
}

func nonEmptyLines(lines []string) []string {
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		kept = append(kept, line)
	}

	return kept
}

func mustMarshalPromptJSON(value any) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "{}"
	}

	return string(data)
}
