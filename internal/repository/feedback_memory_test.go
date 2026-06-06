package repository_test

import (
	"errors"
	"testing"

	"speakmate/internal/model"
	"speakmate/internal/repository"
)

func TestMemoryFeedbackRepositoryFindsCorrectionByMessageID(t *testing.T) {
	repo := repository.NewMemoryFeedbackRepository()
	correction := model.CorrectionResult{
		MessageID:     10,
		SessionID:     1,
		OriginalText:  "I am study computer science.",
		CorrectedText: "I am studying computer science.",
		Errors: []model.CorrectionError{
			{
				Type:        model.CorrectionErrorTypeGrammar,
				Span:        "am study",
				Suggestion:  "am studying",
				Explanation: "be 动词后应接现在分词。",
			},
		},
		BetterExpressions: []string{"I major in computer science."},
	}

	if err := repo.SaveCorrection(correction); err != nil {
		t.Fatalf("SaveCorrection returned error: %v", err)
	}

	found, err := repo.FindCorrectionByMessageID(10)
	if err != nil {
		t.Fatalf("FindCorrectionByMessageID returned error: %v", err)
	}

	if found.MessageID != 10 {
		t.Fatalf("message id = %d, want 10", found.MessageID)
	}
	if found.SessionID != 1 {
		t.Fatalf("session id = %d, want 1", found.SessionID)
	}
	if found.CorrectedText != "I am studying computer science." {
		t.Fatalf("corrected text = %q, want corrected expression", found.CorrectedText)
	}
	if len(found.Errors) != 1 {
		t.Fatalf("errors length = %d, want 1", len(found.Errors))
	}
	if found.Errors[0].Suggestion != "am studying" {
		t.Fatalf("suggestion = %q, want am studying", found.Errors[0].Suggestion)
	}
	if len(found.BetterExpressions) != 1 {
		t.Fatalf("better expressions length = %d, want 1", len(found.BetterExpressions))
	}
}

func TestMemoryFeedbackRepositoryListsCorrectionsBySessionID(t *testing.T) {
	repo := repository.NewMemoryFeedbackRepository()
	corrections := []model.CorrectionResult{
		{MessageID: 1, SessionID: 7, CorrectedText: "first"},
		{MessageID: 2, SessionID: 99, CorrectedText: "other session"},
		{MessageID: 3, SessionID: 7, CorrectedText: "second"},
	}
	for _, correction := range corrections {
		if err := repo.SaveCorrection(correction); err != nil {
			t.Fatalf("SaveCorrection returned error: %v", err)
		}
	}

	found, err := repo.ListCorrectionsBySessionID(7)
	if err != nil {
		t.Fatalf("ListCorrectionsBySessionID returned error: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("corrections length = %d, want 2", len(found))
	}
	if found[0].MessageID != 1 {
		t.Fatalf("first message id = %d, want 1", found[0].MessageID)
	}
	if found[1].MessageID != 3 {
		t.Fatalf("second message id = %d, want 3", found[1].MessageID)
	}
}

func TestMemoryFeedbackRepositoryReturnsCurrentScoreBySessionID(t *testing.T) {
	repo := repository.NewMemoryFeedbackRepository()
	scores := []model.ScoreResult{
		{MessageID: 1, SessionID: 7, TotalScore: 70, Grammar: 68},
		{MessageID: 2, SessionID: 99, TotalScore: 91, Grammar: 90},
		{MessageID: 3, SessionID: 7, TotalScore: 82, Grammar: 80},
	}
	for _, score := range scores {
		if err := repo.SaveScore(score); err != nil {
			t.Fatalf("SaveScore returned error: %v", err)
		}
	}

	current, err := repo.FindCurrentScoreBySessionID(7)
	if err != nil {
		t.Fatalf("FindCurrentScoreBySessionID returned error: %v", err)
	}

	if current.MessageID != 3 {
		t.Fatalf("current message id = %d, want latest score message 3", current.MessageID)
	}
	if current.TotalScore != 82 {
		t.Fatalf("total score = %d, want 82", current.TotalScore)
	}
	if current.Grammar != 80 {
		t.Fatalf("grammar = %d, want 80", current.Grammar)
	}
}

func TestMemoryFeedbackRepositoryReturnsNotFoundErrors(t *testing.T) {
	repo := repository.NewMemoryFeedbackRepository()

	if _, err := repo.FindCorrectionByMessageID(404); !errors.Is(err, repository.ErrCorrectionNotFound) {
		t.Fatalf("FindCorrectionByMessageID error = %v, want ErrCorrectionNotFound", err)
	}
	if _, err := repo.ListCorrectionsBySessionID(404); !errors.Is(err, repository.ErrCorrectionNotFound) {
		t.Fatalf("ListCorrectionsBySessionID error = %v, want ErrCorrectionNotFound", err)
	}
	if _, err := repo.FindCurrentScoreBySessionID(404); !errors.Is(err, repository.ErrScoreNotFound) {
		t.Fatalf("FindCurrentScoreBySessionID error = %v, want ErrScoreNotFound", err)
	}
}

func TestMemoryFeedbackRepositoryReturnsCopies(t *testing.T) {
	repo := repository.NewMemoryFeedbackRepository()
	correction := model.CorrectionResult{
		MessageID:         10,
		SessionID:         1,
		Errors:            []model.CorrectionError{{Suggestion: "am studying"}},
		BetterExpressions: []string{"I major in computer science."},
	}
	if err := repo.SaveCorrection(correction); err != nil {
		t.Fatalf("SaveCorrection returned error: %v", err)
	}

	found, err := repo.FindCorrectionByMessageID(10)
	if err != nil {
		t.Fatalf("FindCorrectionByMessageID returned error: %v", err)
	}
	found.Errors[0].Suggestion = "mutated"
	found.BetterExpressions[0] = "mutated"

	again, err := repo.FindCorrectionByMessageID(10)
	if err != nil {
		t.Fatalf("FindCorrectionByMessageID returned error: %v", err)
	}
	if again.Errors[0].Suggestion != "am studying" {
		t.Fatalf("stored suggestion = %q, want am studying", again.Errors[0].Suggestion)
	}
	if again.BetterExpressions[0] != "I major in computer science." {
		t.Fatalf("stored better expression = %q, want original", again.BetterExpressions[0])
	}
}
