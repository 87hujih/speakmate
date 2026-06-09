# History Insights Design

## Purpose

History currently behaves like a paginated record list. The left summary card in
`web/src/pages/HistoryPage.tsx` calculates an average score from only the current
page of records, so the number changes when the user pages through history and
does not represent real learning progress.

This design adds a dedicated learning insights API and upgrades the history page
into a progress and review surface. The paginated history list remains the place
for session details. The new insights API is responsible for stable, cross-page
learning signals: score trends, scenario trends, frequent errors, and the next
recommended practice.

## Goals

- Show stable 7-day and 30-day learning progress instead of current-page
  statistics.
- Show scenario-specific progress so mixed scenarios do not dilute the meaning
  of the score.
- Surface repeated errors from generated reports as a small "frequent errors"
  library.
- Give the user an obvious next practice recommendation from the history page.
- Keep the existing history list available as training details, with a repeat
  practice entry point.

## Non-Goals

- No long-term spaced repetition scheduler.
- No mastery state per error.
- No new authentication or user identity model.
- No database schema change for the first version.
- No attempt to regenerate or rewrite existing reports.

## API

Add:

```http
GET /api/v1/history/insights?days=30&user_id=42
```

Query parameters:

- `days`: optional. Supported values are `7` and `30`. Default is `30`.
- `user_id`: optional. Uses the same behavior as the existing history list:
  absent means all sessions, present means sessions for that user.

Invalid `days` or `user_id` values return the same history request error style
used by the existing history API.

Response:

```json
{
  "summary": {
    "days": 30,
    "total_sessions": 18,
    "finished_sessions": 15,
    "generated_reports": 12,
    "average_score": 79,
    "previous_average_score": 73,
    "score_delta": 6,
    "best_scenario_code": "restaurant",
    "weakest_scenario_code": "interview"
  },
  "score_trend": [
    {
      "date": "2026-06-01",
      "average_score": 72,
      "session_count": 2
    }
  ],
  "scenario_trends": [
    {
      "scenario": {
        "id": 1,
        "code": "interview",
        "name": "英语面试",
        "difficulty": "medium"
      },
      "session_count": 7,
      "average_score": 76,
      "first_score": 68,
      "latest_score": 82,
      "score_delta": 14,
      "last_trained_at": "2026-06-08T10:20:00Z"
    }
  ],
  "frequent_errors": [
    {
      "category": "grammar",
      "title": "动词时态不稳定",
      "count": 4,
      "latest_original": "I have did a project",
      "latest_suggestion": "I have done a project",
      "scenario_codes": ["interview"],
      "last_seen_at": "2026-06-08T10:20:00Z"
    }
  ],
  "next_recommendation": {
    "type": "scenario_repractice",
    "scenario_id": 1,
    "scenario_code": "interview",
    "scenario_name": "英语面试",
    "reason": "最近 30 天该场景分数提升明显，但语法错误仍重复出现。",
    "focus": "动词时态和项目经历表达",
    "source_session_id": 123
  }
}
```

When no recommendation can be made, `next_recommendation` is `null`.

`next_recommendation` is a small union. For a new same-scenario practice, the
type is `scenario_repractice` and the response includes `scenario_id`,
`scenario_code`, `scenario_name`, `reason`, `focus`, and `source_session_id`.
For an unfinished session, the type is `continue_session` and the response
includes `session_id`, `scenario_code`, `scenario_name`, `reason`, and `focus`.

## Backend Design

Add a dedicated insights service rather than extending the paginated list
service:

- `internal/service/history_insights.go`
- `internal/handler/history_insights.go`

Router wiring in `internal/router/router.go`:

```go
api.GET("/history/insights", historyInsightHandler.Get)
```

The new service depends on existing repositories where possible:

- Session repository: list sessions by user and time window.
- Scenario reader: load scenario summaries for scenario trend rows.
- Feedback repository: read the current score for each session.
- Report repository: read generated reports for frequent errors and practice
  recommendations.

The current session repository only supports paginated listing. For this API,
add a narrow repository method for windowed history reads instead of requesting
many pages from the service layer:

```go
ListSessionsByWindow(query model.SessionWindowQuery) ([]model.Session, error)
```

`SessionWindowQuery` includes:

- `UserID int`
- `StartedAt time.Time`
- `EndedAt time.Time`
- `Limit int`

The first version should cap `Limit` at a conservative value such as `500`.
This avoids unbounded scans while still being enough for a 30-day personal
learning dashboard.

## Insight Rules

Time windows:

- Current window: now minus `days` through now.
- Previous window: the same length immediately before the current window.
- `days=7` compares the last 7 days with the previous 7 days.
- `days=30` compares the last 30 days with the previous 30 days.
- A session belongs to a window by `training_sessions.created_at`, not
  `ended_at`.
- Daily grouping uses the backend server's local timezone for the first version.
  A user timezone parameter can be added later if the product adds accounts and
  locale settings.

Score inclusion:

- Only sessions with a score participate in averages and trends.
- Running sessions count toward `total_sessions` but not score averages.
- Finished sessions without a score count toward completion statistics but not
  score averages.

Summary:

- `total_sessions`: all sessions in the current window.
- `finished_sessions`: current-window sessions with status `finished`.
- `generated_reports`: current-window sessions with an existing report.
- `average_score`: rounded average across scored current-window sessions.
- `previous_average_score`: rounded average across scored previous-window
  sessions.
- `score_delta`: `average_score - previous_average_score`. If either side has
  no scored data, return `0`.
- `best_scenario_code`: scenario with the highest current-window average score.
- `weakest_scenario_code`: scenario with the lowest current-window average
  score among scenarios that have scored sessions.

Score trend:

- Group current-window scored sessions by local calendar date.
- Each point contains rounded average score and scored session count.
- Dates without scored sessions are omitted.

Scenario trends:

- Group current-window scored sessions by `scenario_id`.
- `first_score` is the oldest scored session in the current window.
- `latest_score` is the newest scored session in the current window.
- `score_delta` is `latest_score - first_score`.
- Sort by most recent training time descending.

Frequent errors:

- Use only generated reports in the current window.
- Aggregate `frequent_errors_json` into at most 5 rows.
- Normalize each error key by lowercasing, trimming whitespace, and preferring
  the parsed correction title when available. If only a plain string exists,
  use the first stable phrase before `->` as the key.
- Store latest evidence fields from the newest report where the error appeared.
- Sort by `count` descending, then `last_seen_at` descending.

Recommendation:

Return one recommendation with this priority:

1. If a frequent error appears at least twice, recommend re-practicing the
   newest scenario where that error appeared.
2. Otherwise, recommend the current-window scored scenario with the lowest
   average score.
3. Otherwise, recommend continuing the newest running session.
4. Otherwise, return `null`.

The recommendation is not responsible for creating a session. For
`scenario_repractice`, the frontend uses the scenario id to call the existing
session creation flow. For `continue_session`, the frontend links to the
existing training session.

## Frontend Design

Add backend client and adapter types:

- `BackendHistoryInsights`
- `BackendHistoryInsightSummary`
- `BackendHistoryScoreTrendPoint`
- `BackendScenarioTrend`
- `BackendFrequentErrorInsight`
- `BackendNextPracticeRecommendation`

Add frontend model types:

- `HistoryInsights`
- `HistoryInsightSummary`
- `HistoryScoreTrendPoint`
- `ScenarioTrend`
- `FrequentErrorInsight`
- `NextPracticeRecommendation`

Add loader:

```ts
export async function loadHistoryInsights(days = 30, client = apiClient) {
  return mapHistoryInsights(await client.getHistoryInsights(days));
}
```

`HistoryPage` loads the paginated records and insights independently. If the
insights request fails, the history list still renders; the insights area shows
an error state with a retry button.

Recommended page structure:

- Header: keep "开始新训练".
- Top insight band: next recommended practice with a primary action.
- Left summary panel: 7/30 day average, score delta, finished sessions, report
  count. Remove current-page average score.
- Main insight area:
  - 7/30 day segmented control.
  - Score trend chart.
  - Scenario trend cards.
  - Frequent error cards.
- Training details: existing paginated history list.

History cards should add a repeat action for finished sessions:

- Existing main action remains "查看报告" or "生成报告".
- Add secondary action "再练一次同场景".
- Running sessions keep "继续训练" as the primary action.

## Empty and Error States

No history:

- Summary values show zero or `--` where a score is not meaningful.
- Trend and scenario sections show a short empty state.
- Recommendation is absent.
- The existing empty history card remains available.

No reports:

- Frequent errors show an empty state explaining that reports are needed for
  error aggregation.
- Score trend and scenario trend still work if scores exist.

Insights load failure:

- Do not block history list rendering.
- Show a compact retry affordance in the insights area.

## Testing

Backend service tests in `internal/service/history_insights_test.go`:

- Empty history returns empty insight arrays and `next_recommendation: null`.
- Running sessions count in totals but do not affect averages.
- Current and previous windows calculate averages and deltas correctly.
- `days=7` and `days=30` filter sessions correctly.
- Scenario trend `first_score`, `latest_score`, and `score_delta` are correct.
- Frequent errors are normalized, counted, sorted, and capped at 5.
- Recommendation priority chooses repeated error, then weakest scenario, then
  running session, then null.

Backend handler/router tests:

- `GET /api/v1/history/insights` returns success with the expected response
  shape.
- Invalid `days` returns the history invalid request error.
- Optional `user_id` filters results consistently with the history list.

Frontend tests:

- `web/src/api/adapters.test.ts` covers `mapHistoryInsights`.
- `web/src/api/loaders.test.ts` covers `loadHistoryInsights`.
- `web/src/pages/HistoryPage.layout.test.tsx` covers:
  - next recommendation rendering,
  - summary uses insight average instead of current-page average,
  - insights error does not hide the history list,
  - empty insights state.

## Rollout

Build in one focused MVP:

1. Backend insights API and tests.
2. Frontend client, adapter, loader, and tests.
3. History page insight sections and repeat-practice action.

The existing `/api/v1/sessions` history list remains unchanged so older callers
do not break.
