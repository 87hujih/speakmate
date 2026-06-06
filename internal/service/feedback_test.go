package service_test

import (
	"errors"
	"testing"

	"speakmate/internal/model"
	"speakmate/internal/repository"
	"speakmate/internal/service"
)

func TestFeedbackServiceGetsMessageCorrection(t *testing.T) {
	repo := newFakeFeedbackRepository()
	repo.correctionsByMessageID[10] = model.CorrectionResult{
		MessageID:     10,
		SessionID:     1,
		CorrectedText: "I am studying computer science.",
	}
	feedbackService := service.NewFeedbackService(repo)

	correction, err := feedbackService.GetMessageCorrection(10)
	if err != nil {
		t.Fatalf("GetMessageCorrection returned error: %v", err)
	}

	if correction.MessageID != 10 {
		t.Fatalf("message id = %d, want 10", correction.MessageID)
	}
	if correction.CorrectedText != "I am studying computer science." {
		t.Fatalf("corrected text = %q, want saved correction", correction.CorrectedText)
	}
}

func TestFeedbackServiceListsSessionCorrections(t *testing.T) {
	repo := newFakeFeedbackRepository()
	repo.correctionsBySessionID[1] = []model.CorrectionResult{
		{MessageID: 10, SessionID: 1},
		{MessageID: 12, SessionID: 1},
	}
	feedbackService := service.NewFeedbackService(repo)

	corrections, err := feedbackService.ListSessionCorrections(1)
	if err != nil {
		t.Fatalf("ListSessionCorrections returned error: %v", err)
	}

	if len(corrections) != 2 {
		t.Fatalf("corrections length = %d, want 2", len(corrections))
	}
	if corrections[0].MessageID != 10 {
		t.Fatalf("first message id = %d, want 10", corrections[0].MessageID)
	}
	if corrections[1].MessageID != 12 {
		t.Fatalf("second message id = %d, want 12", corrections[1].MessageID)
	}
}

func TestFeedbackServiceGetsSessionCurrentScore(t *testing.T) {
	repo := newFakeFeedbackRepository()
	repo.scoresBySessionID[1] = model.ScoreResult{
		MessageID:  12,
		SessionID:  1,
		Grammar:    80,
		TotalScore: 82,
	}
	feedbackService := service.NewFeedbackService(repo)

	score, err := feedbackService.GetSessionCurrentScore(1)
	if err != nil {
		t.Fatalf("GetSessionCurrentScore returned error: %v", err)
	}

	if score.MessageID != 12 {
		t.Fatalf("message id = %d, want 12", score.MessageID)
	}
	if score.TotalScore != 82 {
		t.Fatalf("total score = %d, want 82", score.TotalScore)
	}
}

func TestFeedbackServiceMapsRepositoryNotFoundErrors(t *testing.T) {
	repo := newFakeFeedbackRepository()
	feedbackService := service.NewFeedbackService(repo)

	if _, err := feedbackService.GetMessageCorrection(10); !errors.Is(err, service.ErrCorrectionNotFound) {
		t.Fatalf("GetMessageCorrection error = %v, want ErrCorrectionNotFound", err)
	}
	if _, err := feedbackService.ListSessionCorrections(1); !errors.Is(err, service.ErrCorrectionNotFound) {
		t.Fatalf("ListSessionCorrections error = %v, want ErrCorrectionNotFound", err)
	}
	if _, err := feedbackService.GetSessionCurrentScore(1); !errors.Is(err, service.ErrScoreNotFound) {
		t.Fatalf("GetSessionCurrentScore error = %v, want ErrScoreNotFound", err)
	}
}

func TestFeedbackServiceRejectsInvalidIDs(t *testing.T) {
	repo := newFakeFeedbackRepository()
	feedbackService := service.NewFeedbackService(repo)

	if _, err := feedbackService.GetMessageCorrection(0); !errors.Is(err, service.ErrInvalidFeedbackRequest) {
		t.Fatalf("GetMessageCorrection error = %v, want ErrInvalidFeedbackRequest", err)
	}
	if _, err := feedbackService.ListSessionCorrections(0); !errors.Is(err, service.ErrInvalidFeedbackRequest) {
		t.Fatalf("ListSessionCorrections error = %v, want ErrInvalidFeedbackRequest", err)
	}
	if _, err := feedbackService.GetSessionCurrentScore(0); !errors.Is(err, service.ErrInvalidFeedbackRequest) {
		t.Fatalf("GetSessionCurrentScore error = %v, want ErrInvalidFeedbackRequest", err)
	}
	if repo.callCount != 0 {
		t.Fatalf("repo call count = %d, want 0", repo.callCount)
	}
}

type fakeFeedbackRepository struct {
	callCount              int
	correctionsByMessageID map[int]model.CorrectionResult
	correctionsBySessionID map[int][]model.CorrectionResult
	scoresBySessionID      map[int]model.ScoreResult
	err                    error
}

func newFakeFeedbackRepository() *fakeFeedbackRepository {
	return &fakeFeedbackRepository{
		correctionsByMessageID: make(map[int]model.CorrectionResult),
		correctionsBySessionID: make(map[int][]model.CorrectionResult),
		scoresBySessionID:      make(map[int]model.ScoreResult),
	}
}

func (r *fakeFeedbackRepository) SaveCorrection(correction model.CorrectionResult) error {
	r.callCount++
	if r.err != nil {
		return r.err
	}
	r.correctionsByMessageID[correction.MessageID] = correction
	r.correctionsBySessionID[correction.SessionID] = append(r.correctionsBySessionID[correction.SessionID], correction)

	return nil
}

func (r *fakeFeedbackRepository) SaveScore(score model.ScoreResult) error {
	r.callCount++
	if r.err != nil {
		return r.err
	}
	r.scoresBySessionID[score.SessionID] = score

	return nil
}

func (r *fakeFeedbackRepository) FindCorrectionByMessageID(messageID int) (model.CorrectionResult, error) {
	r.callCount++
	if r.err != nil {
		return model.CorrectionResult{}, r.err
	}
	correction, ok := r.correctionsByMessageID[messageID]
	if !ok {
		return model.CorrectionResult{}, repository.ErrCorrectionNotFound
	}

	return correction, nil
}

func (r *fakeFeedbackRepository) ListCorrectionsBySessionID(sessionID int) ([]model.CorrectionResult, error) {
	r.callCount++
	if r.err != nil {
		return nil, r.err
	}
	corrections, ok := r.correctionsBySessionID[sessionID]
	if !ok {
		return nil, repository.ErrCorrectionNotFound
	}

	return corrections, nil
}

func (r *fakeFeedbackRepository) FindCurrentScoreBySessionID(sessionID int) (model.ScoreResult, error) {
	r.callCount++
	if r.err != nil {
		return model.ScoreResult{}, r.err
	}
	score, ok := r.scoresBySessionID[sessionID]
	if !ok {
		return model.ScoreResult{}, repository.ErrScoreNotFound
	}

	return score, nil
}
