package service

import (
	"errors"
	"time"

	"speakmate/internal/model"
	"speakmate/internal/repository"
)

var (
	// ErrInvalidSessionRequest 表示创建 Session 的业务参数非法。
	ErrInvalidSessionRequest = errors.New("invalid session request")
	// ErrSessionNotFound 表示业务层没有找到对应 Session。
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionAlreadyFinished 表示 Session 已经结束，不能执行运行中操作。
	ErrSessionAlreadyFinished = errors.New("session already finished")
)

// ScenarioReader 定义 Session 服务依赖的场景读取能力。
type ScenarioReader interface {
	GetScenario(id int) (model.Scenario, error)
}

// SessionRepository 定义 Session 服务依赖的数据访问能力。
type SessionRepository interface {
	Create(session model.Session) (model.Session, error)
	FindByID(id int) (model.Session, error)
	Finish(id int, endedAt time.Time) (model.Session, error)
}

// SessionService 封装训练 Session 生命周期业务流程。
type SessionService struct {
	scenarioReader ScenarioReader
	repo           SessionRepository
	now            func() time.Time
}

// NewSessionService 创建 Session 服务实例。
func NewSessionService(scenarioReader ScenarioReader, repo SessionRepository) *SessionService {
	return &SessionService{
		scenarioReader: scenarioReader,
		repo:           repo,
		now:            time.Now,
	}
}

// CreateSessionInput 是创建训练 Session 的业务输入。
type CreateSessionInput struct {
	ScenarioID int
	UserID     int
}

// CreateSessionResult 是创建训练 Session 的业务输出。
type CreateSessionResult struct {
	Session        model.Session
	OpeningMessage string
}

// GetSessionResult 是查询训练 Session 的业务输出。
type GetSessionResult struct {
	Session  model.Session
	Scenario model.Scenario
}

// CreateSession 基于有效场景创建 running 状态的训练 Session。
func (s *SessionService) CreateSession(input CreateSessionInput) (CreateSessionResult, error) {
	if input.ScenarioID <= 0 {
		return CreateSessionResult{}, ErrInvalidSessionRequest
	}
	if input.UserID <= 0 {
		input.UserID = 1
	}

	scenario, err := s.scenarioReader.GetScenario(input.ScenarioID)
	if err != nil {
		if errors.Is(err, ErrScenarioNotFound) {
			return CreateSessionResult{}, ErrScenarioNotFound
		}

		return CreateSessionResult{}, err
	}

	session := model.Session{
		ScenarioID: input.ScenarioID,
		UserID:     input.UserID,
		Status:     model.SessionStatusRunning,
		TurnCount:  0,
		Messages:   []model.Message{},
		CreatedAt:  s.now(),
	}
	created, err := s.repo.Create(session)
	if err != nil {
		return CreateSessionResult{}, err
	}

	return CreateSessionResult{
		Session:        created,
		OpeningMessage: scenario.OpeningMessage,
	}, nil
}

// GetSession 查询 Session，并补齐关联场景信息。
func (s *SessionService) GetSession(id int) (GetSessionResult, error) {
	session, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return GetSessionResult{}, ErrSessionNotFound
		}

		return GetSessionResult{}, err
	}

	scenario, err := s.scenarioReader.GetScenario(session.ScenarioID)
	if err != nil {
		return GetSessionResult{}, err
	}

	return GetSessionResult{
		Session:  session,
		Scenario: scenario,
	}, nil
}

// FinishSession 结束 running 状态的 Session。
func (s *SessionService) FinishSession(id int) (model.Session, error) {
	session, err := s.repo.Finish(id, s.now())
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return model.Session{}, ErrSessionNotFound
		}
		if errors.Is(err, repository.ErrSessionAlreadyFinished) {
			return model.Session{}, ErrSessionAlreadyFinished
		}

		return model.Session{}, err
	}

	return session, nil
}
