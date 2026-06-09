package repository_test

import (
	"testing"
	"time"

	"speakmate/internal/model"
	"speakmate/internal/repository"
)

func TestMemorySessionRepositoryListSessionsByWindow(t *testing.T) {
	base := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	repo := repository.NewMemorySessionRepository()

	beforeStart := mustCreateMemorySession(t, repo, model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusFinished,
		TurnCount:  1,
		CreatedAt:  base.Add(-time.Minute),
	})
	atStart := mustCreateMemorySession(t, repo, model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusFinished,
		TurnCount:  1,
		Messages: []model.Message{{
			Role:      model.MessageRoleUser,
			Content:   "original",
			CreatedAt: base,
		}},
		CreatedAt: base,
	})
	insideSameTimeLowID := mustCreateMemorySession(t, repo, model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusFinished,
		TurnCount:  1,
		CreatedAt:  base.Add(time.Hour),
	})
	insideSameTimeHighID := mustCreateMemorySession(t, repo, model.Session{
		ScenarioID: 1,
		UserID:     2,
		Status:     model.SessionStatusFinished,
		TurnCount:  1,
		CreatedAt:  base.Add(time.Hour),
	})
	newest := mustCreateMemorySession(t, repo, model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusFinished,
		TurnCount:  1,
		CreatedAt:  base.Add(2 * time.Hour),
	})
	atEnd := mustCreateMemorySession(t, repo, model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusFinished,
		TurnCount:  1,
		CreatedAt:  base.Add(3 * time.Hour),
	})

	t.Run("filters by inclusive start and exclusive end", func(t *testing.T) {
		sessions, err := repo.ListSessionsByWindow(model.SessionWindowQuery{
			StartedAt: base,
			EndedAt:   base.Add(3 * time.Hour),
		})
		if err != nil {
			t.Fatalf("ListSessionsByWindow returned error: %v", err)
		}

		assertSessionIDs(t, sessions, []int{
			newest.ID,
			insideSameTimeHighID.ID,
			insideSameTimeLowID.ID,
			atStart.ID,
		})
		assertSessionIDAbsent(t, sessions, beforeStart.ID)
		assertSessionIDAbsent(t, sessions, atEnd.ID)
	})

	t.Run("applies optional user filter", func(t *testing.T) {
		sessions, err := repo.ListSessionsByWindow(model.SessionWindowQuery{
			UserID:    1,
			StartedAt: base,
			EndedAt:   base.Add(3 * time.Hour),
		})
		if err != nil {
			t.Fatalf("ListSessionsByWindow returned error: %v", err)
		}

		assertSessionIDs(t, sessions, []int{
			newest.ID,
			insideSameTimeLowID.ID,
			atStart.ID,
		})
	})

	t.Run("caps results by limit", func(t *testing.T) {
		sessions, err := repo.ListSessionsByWindow(model.SessionWindowQuery{
			StartedAt: base,
			EndedAt:   base.Add(3 * time.Hour),
			Limit:     2,
		})
		if err != nil {
			t.Fatalf("ListSessionsByWindow returned error: %v", err)
		}

		assertSessionIDs(t, sessions, []int{
			newest.ID,
			insideSameTimeHighID.ID,
		})
	})

	t.Run("returns cloned sessions", func(t *testing.T) {
		sessions, err := repo.ListSessionsByWindow(model.SessionWindowQuery{
			UserID:    1,
			StartedAt: base,
			EndedAt:   base.Add(time.Minute),
			Limit:     1,
		})
		if err != nil {
			t.Fatalf("ListSessionsByWindow returned error: %v", err)
		}
		if len(sessions) != 1 {
			t.Fatalf("sessions length = %d, want 1", len(sessions))
		}
		sessions[0].Messages[0].Content = "mutated"

		found, err := repo.FindByID(atStart.ID)
		if err != nil {
			t.Fatalf("FindByID returned error: %v", err)
		}
		if found.Messages[0].Content != "original" {
			t.Fatalf("stored message content = %q, want original", found.Messages[0].Content)
		}
	})
}

func mustCreateMemorySession(t *testing.T, repo *repository.MemorySessionRepository, session model.Session) model.Session {
	t.Helper()

	created, err := repo.Create(session)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	return created
}

func assertSessionIDs(t *testing.T, sessions []model.Session, want []int) {
	t.Helper()

	if len(sessions) != len(want) {
		t.Fatalf("sessions length = %d, want %d; sessions = %#v", len(sessions), len(want), sessions)
	}
	for i, session := range sessions {
		if session.ID != want[i] {
			t.Fatalf("sessions[%d].ID = %d, want %d; sessions = %#v", i, session.ID, want[i], sessions)
		}
	}
}

func assertSessionIDAbsent(t *testing.T, sessions []model.Session, id int) {
	t.Helper()

	for _, session := range sessions {
		if session.ID == id {
			t.Fatalf("session id %d unexpectedly present in %#v", id, sessions)
		}
	}
}
