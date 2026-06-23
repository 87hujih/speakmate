package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"speakmate/internal/agent"
	"speakmate/internal/model"
	"speakmate/internal/repository"
	"speakmate/internal/state"
	"speakmate/internal/stream"
)

// 服务层复用的哨兵错误。
var (
	// ErrInvalidSessionRequest 表示创建 Session 的业务参数非法。
	ErrInvalidSessionRequest = errors.New("训练请求无效")
	// ErrSessionNotFound 表示业务层没有找到对应 Session。
	ErrSessionNotFound = errors.New("未找到训练")
	// ErrSessionAlreadyFinished 表示 Session 已经结束，不能执行运行中操作。
	ErrSessionAlreadyFinished = errors.New("训练已结束")
	// ErrInvalidMessageRequest 表示发送消息的业务参数非法。
	ErrInvalidMessageRequest = errors.New("消息请求无效")
	// ErrMessageContentRequired 表示消息内容不能为空。
	ErrMessageContentRequired = errors.New("消息内容不能为空")
	// ErrConversationAgentFailed 表示对话 Agent 生成回复失败。
	ErrConversationAgentFailed = errors.New("对话 AI 回复失败")
	// ErrFeedbackAgentFailed 表示反馈 Agent 生成纠错或评分失败。
	ErrFeedbackAgentFailed = errors.New("反馈 AI 生成失败")
	// ErrStateStoreFailed 表示短期状态写入失败。
	ErrStateStoreFailed = errors.New("训练短期状态写入失败")
	// ErrEventPublishFailed 表示 SSE/WebSocket 事件发布失败。
	ErrEventPublishFailed = errors.New("实时事件发布失败")
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
	AppendTurn(id int, userMessage model.Message, aiMessage model.Message) (model.Session, error)
}

// EventPublisher 定义业务事件发布能力。
type EventPublisher interface {
	Publish(event stream.Event) error
}

// SessionService 封装训练 Session 生命周期业务流程。
type SessionService struct {
	scenarioReader   ScenarioReader
	repo             SessionRepository
	feedbackRepo     FeedbackRepository
	stateStore       state.SessionStateStore
	events           EventPublisher
	conversation     agent.ConversationAgent
	correction       agent.CorrectionAgent
	scoring          agent.ScoringAgent
	feedbackFailOpen bool
	now              func() time.Time
}

// SessionOption 用于配置 SessionService。
type SessionOption func(*SessionService)

// NewSessionService 创建 Session 服务实例。
func NewSessionService(scenarioReader ScenarioReader, repo SessionRepository, opts ...SessionOption) *SessionService {
	service := &SessionService{
		scenarioReader:   scenarioReader,
		repo:             repo,
		conversation:     agent.NewMockConversationAgent(),
		correction:       agent.NewMockCorrectionAgent(),
		scoring:          agent.NewMockScoringAgent(),
		feedbackFailOpen: true,
		now:              time.Now,
	}
	for _, opt := range opts {
		opt(service)
	}

	return service
}

// WithConversationAgent 返回用于覆盖默认行为的配置选项。
func WithConversationAgent(conversation agent.ConversationAgent) SessionOption {
	return func(service *SessionService) {
		if conversation != nil {
			service.conversation = conversation
		}
	}
}

// WithFeedbackRepository 返回用于覆盖默认行为的配置选项。
func WithFeedbackRepository(feedbackRepo FeedbackRepository) SessionOption {
	return func(service *SessionService) {
		if feedbackRepo != nil {
			service.feedbackRepo = feedbackRepo
		}
	}
}

// WithEventPublisher 返回用于覆盖默认行为的配置选项。
func WithEventPublisher(publisher EventPublisher) SessionOption {
	return func(service *SessionService) {
		if publisher != nil {
			service.events = publisher
		}
	}
}

// WithStateStore 返回用于覆盖默认行为的配置选项。
func WithStateStore(store state.SessionStateStore) SessionOption {
	return func(service *SessionService) {
		if store != nil {
			service.stateStore = store
		}
	}
}

// WithCorrectionAgent 返回用于覆盖默认行为的配置选项。
func WithCorrectionAgent(correction agent.CorrectionAgent) SessionOption {
	return func(service *SessionService) {
		if correction != nil {
			service.correction = correction
		}
	}
}

// WithScoringAgent 返回用于覆盖默认行为的配置选项。
func WithScoringAgent(scoring agent.ScoringAgent) SessionOption {
	return func(service *SessionService) {
		if scoring != nil {
			service.scoring = scoring
		}
	}
}

// WithFeedbackFailOpen 返回用于覆盖默认行为的配置选项。
func WithFeedbackFailOpen(failOpen bool) SessionOption {
	return func(service *SessionService) {
		service.feedbackFailOpen = failOpen
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

// SendMessageInput 是发送文本消息的业务输入。
type SendMessageInput struct {
	SessionID int
	Content   string
	Context   context.Context
}

// SendMessageResult 是发送消息后的业务输出。
type SendMessageResult struct {
	UserMessage       model.Message
	AIMessage         model.Message
	Stage             string
	NextGoal          string
	TurnCount         int
	CorrectionSummary CorrectionSummary
	ScoreSummary      ScoreSummary
}

// CorrectionSummary 是发送消息响应中返回的纠错摘要。
type CorrectionSummary struct {
	HasErrors  bool `json:"has_errors"`
	ErrorCount int  `json:"error_count"`
}

// ScoreSummary 是发送消息响应中返回的评分摘要。
type ScoreSummary struct {
	TotalScore int `json:"total_score"`
	Grammar    int `json:"grammar"`
	Expression int `json:"expression"`
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
	if err := s.saveSessionState(context.Background(), created, ""); err != nil {
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
	if err := s.saveSessionState(context.Background(), session, currentSessionStage(session)); err != nil {
		return model.Session{}, err
	}

	return session, nil
}

// SendMessage 保存用户消息，生成 AI 回复，并推进对话轮次。
func (s *SessionService) SendMessage(input SendMessageInput) (SendMessageResult, error) {
	if input.SessionID <= 0 {
		return SendMessageResult{}, ErrInvalidMessageRequest
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return SendMessageResult{}, ErrMessageContentRequired
	}

	session, err := s.repo.FindByID(input.SessionID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return SendMessageResult{}, ErrSessionNotFound
		}

		return SendMessageResult{}, err
	}
	if session.Status == model.SessionStatusFinished {
		return SendMessageResult{}, ErrSessionAlreadyFinished
	}

	scenario, err := s.scenarioReader.GetScenario(session.ScenarioID)
	if err != nil {
		return SendMessageResult{}, err
	}

	conversation := s.conversation
	if conversation == nil {
		conversation = agent.NewMockConversationAgent()
	}
	ctx := input.Context
	if ctx == nil {
		ctx = context.Background()
	}
	conversationInput := agent.ConversationInput{
		Scenario:    scenario,
		Session:     session,
		History:     session.Messages,
		UserContent: content,
	}
	reply, usedStreaming, err := s.generateConversationReply(ctx, input.SessionID, conversation, conversationInput)
	if err != nil {
		if errors.Is(err, ErrEventPublishFailed) {
			return SendMessageResult{}, err
		}
		wrapped := fmt.Errorf("%w: %v", ErrConversationAgentFailed, err)
		s.publishSessionError(input.SessionID, wrapped)
		return SendMessageResult{}, wrapped
	}
	reply.Reply = strings.TrimSpace(reply.Reply)
	if reply.Reply == "" {
		return SendMessageResult{}, ErrConversationAgentFailed
	}
	if reply.Stage == "" {
		reply.Stage = agent.StageNameForTurn(scenario.Stages, session.TurnCount+1)
	}
	if reply.NextGoal == "" {
		reply.NextGoal = agent.NextGoalForTurn(scenario.Stages, session.TurnCount+1)
	}

	createdAt := s.now()
	userMessage := model.Message{
		Role:      model.MessageRoleUser,
		Content:   content,
		Stage:     agent.StageNameForTurn(scenario.Stages, session.TurnCount),
		CreatedAt: createdAt,
	}
	aiMessage := model.Message{
		Role:      model.MessageRoleAI,
		Content:   reply.Reply,
		Stage:     reply.Stage,
		CreatedAt: createdAt,
	}

	updated, err := s.repo.AppendTurn(session.ID, userMessage, aiMessage)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return SendMessageResult{}, ErrSessionNotFound
		}
		if errors.Is(err, repository.ErrSessionAlreadyFinished) {
			return SendMessageResult{}, ErrSessionAlreadyFinished
		}

		return SendMessageResult{}, err
	}
	if len(updated.Messages) < 2 {
		return SendMessageResult{}, errors.New("追加对话轮次后返回的消息不完整")
	}

	messages := updated.Messages
	savedUserMessage := messages[len(messages)-2]
	savedAIMessage := messages[len(messages)-1]
	if err := s.saveMessageState(ctx, updated, savedAIMessage.Stage); err != nil {
		return SendMessageResult{}, err
	}
	if !usedStreaming {
		if err := s.publishStreamEvent(stream.Event{
			Type:      stream.EventTypeAIMessageDelta,
			SessionID: updated.ID,
			Payload: stream.AIMessageDeltaPayload{
				MessageID: savedAIMessage.ID,
				Delta:     savedAIMessage.Content,
			},
		}); err != nil {
			return SendMessageResult{}, err
		}
	}
	if err := s.publishStreamEvent(stream.Event{
		Type:      stream.EventTypeAIMessageDone,
		SessionID: updated.ID,
		Payload: stream.AIMessageDonePayload{
			MessageID: savedAIMessage.ID,
			Content:   savedAIMessage.Content,
			Stage:     savedAIMessage.Stage,
		},
	}); err != nil {
		return SendMessageResult{}, err
	}
	correctionSummary, scoreSummary, err := s.generateFeedback(ctx, scenario, updated, savedUserMessage)
	if err != nil {
		s.publishSessionError(updated.ID, err)
		return SendMessageResult{}, err
	}

	return SendMessageResult{
		UserMessage:       savedUserMessage,
		AIMessage:         savedAIMessage,
		Stage:             savedAIMessage.Stage,
		NextGoal:          reply.NextGoal,
		TurnCount:         updated.TurnCount,
		CorrectionSummary: correctionSummary,
		ScoreSummary:      scoreSummary,
	}, nil
}

// generateConversationReply 生成 AI 回复并发布流式事件。
func (s *SessionService) generateConversationReply(ctx context.Context, sessionID int, conversation agent.ConversationAgent, input agent.ConversationInput) (agent.ConversationOutput, bool, error) {
	if streamingConversation, ok := conversation.(agent.StreamingConversationAgent); ok {
		reply, err := streamingConversation.StreamReply(ctx, input, func(delta agent.ConversationDelta) error {
			if delta.Content == "" {
				return nil
			}
			if err := s.publishStreamEvent(stream.Event{
				Type:      stream.EventTypeAIMessageDelta,
				SessionID: sessionID,
				Payload: stream.AIMessageDeltaPayload{
					MessageID: 0,
					Delta:     delta.Content,
				},
			}); err != nil {
				return err
			}

			return nil
		})

		return reply, true, err
	}

	reply, err := conversation.GenerateReply(ctx, input)
	return reply, false, err
}

// generateFeedback 为用户消息生成纠错和评分。
func (s *SessionService) generateFeedback(ctx context.Context, scenario model.Scenario, session model.Session, userMessage model.Message) (CorrectionSummary, ScoreSummary, error) {
	if s.feedbackRepo == nil {
		return CorrectionSummary{}, ScoreSummary{}, nil
	}

	correctionAgent := s.correction
	if correctionAgent == nil {
		correctionAgent = agent.NewMockCorrectionAgent()
	}
	correctionOutput, err := correctionAgent.Correct(agent.CorrectionInput{
		Scenario:    scenario,
		Session:     session,
		History:     session.Messages,
		UserMessage: userMessage,
	})
	if err != nil {
		return s.handleFeedbackFailure(CorrectionSummary{}, ScoreSummary{}, "纠错 Agent 失败", err)
	}
	correction := normalizeCorrectionResult(correctionOutput.Result, session.ID, userMessage)
	if err := s.feedbackRepo.SaveCorrection(correction); err != nil {
		return s.handleFeedbackFailure(CorrectionSummary{}, ScoreSummary{}, "保存纠错结果失败", err)
	}
	if err := s.appendCorrectionState(ctx, correction); err != nil {
		return CorrectionSummary{}, ScoreSummary{}, err
	}
	correctionSummary := correctionSummaryFromResult(correction)
	if err := s.publishStreamEvent(stream.Event{
		Type:      stream.EventTypeCorrectionDone,
		SessionID: session.ID,
		Payload: stream.CorrectionDonePayload{
			MessageID:  correction.MessageID,
			HasErrors:  correctionSummary.HasErrors,
			ErrorCount: correctionSummary.ErrorCount,
		},
	}); err != nil {
		return CorrectionSummary{}, ScoreSummary{}, err
	}

	scoringAgent := s.scoring
	if scoringAgent == nil {
		scoringAgent = agent.NewMockScoringAgent()
	}
	scoreOutput, err := scoringAgent.Score(agent.ScoringInput{
		Scenario:    scenario,
		Session:     session,
		History:     session.Messages,
		UserMessage: userMessage,
		Correction:  correction,
	})
	if err != nil {
		return s.handleFeedbackFailure(correctionSummaryFromResult(correction), ScoreSummary{}, "评分 Agent 失败", err)
	}
	score := normalizeScoreResult(scoreOutput.Result, session.ID, userMessage, correction)
	if err := s.feedbackRepo.SaveScore(score); err != nil {
		return s.handleFeedbackFailure(correctionSummary, ScoreSummary{}, "保存评分结果失败", err)
	}
	if err := s.savePartialScoreState(ctx, score); err != nil {
		return CorrectionSummary{}, ScoreSummary{}, err
	}
	scoreSummary := scoreSummaryFromResult(score)
	if err := s.publishStreamEvent(stream.Event{
		Type:      stream.EventTypeScoreUpdated,
		SessionID: session.ID,
		Payload: stream.ScoreUpdatedPayload{
			MessageID:  score.MessageID,
			TotalScore: scoreSummary.TotalScore,
			Grammar:    scoreSummary.Grammar,
			Expression: scoreSummary.Expression,
		},
	}); err != nil {
		return CorrectionSummary{}, ScoreSummary{}, err
	}

	return correctionSummary, scoreSummary, nil
}

// handleFeedbackFailure 根据 fail-open 策略处理反馈生成失败。
func (s *SessionService) handleFeedbackFailure(correctionSummary CorrectionSummary, scoreSummary ScoreSummary, step string, err error) (CorrectionSummary, ScoreSummary, error) {
	if s.feedbackFailOpen {
		return correctionSummary, scoreSummary, nil
	}

	return CorrectionSummary{}, ScoreSummary{}, fmt.Errorf("%w：%s：%v", ErrFeedbackAgentFailed, step, err)
}

// normalizeCorrectionResult 归一化纠错结果的消息和 Session 归属。
func normalizeCorrectionResult(correction model.CorrectionResult, sessionID int, userMessage model.Message) model.CorrectionResult {
	if correction.MessageID == 0 {
		correction.MessageID = userMessage.ID
	}
	if correction.SessionID == 0 {
		correction.SessionID = sessionID
	}
	if correction.OriginalText == "" {
		correction.OriginalText = userMessage.Content
	}
	if correction.CorrectedText == "" {
		correction.CorrectedText = correction.OriginalText
	}
	if correction.Errors == nil {
		correction.Errors = []model.CorrectionError{}
	}
	if correction.BetterExpressions == nil {
		correction.BetterExpressions = []string{}
	}

	return correction
}

// normalizeScoreResult 归一化评分结果的消息和 Session 归属。
func normalizeScoreResult(score model.ScoreResult, sessionID int, userMessage model.Message, correction model.CorrectionResult) model.ScoreResult {
	if score.MessageID == 0 {
		score.MessageID = correction.MessageID
	}
	if score.MessageID == 0 {
		score.MessageID = userMessage.ID
	}
	if score.SessionID == 0 {
		score.SessionID = correction.SessionID
	}
	if score.SessionID == 0 {
		score.SessionID = sessionID
	}

	return score
}

// correctionSummaryFromResult 从纠错结果生成摘要。
func correctionSummaryFromResult(correction model.CorrectionResult) CorrectionSummary {
	errorCount := len(correction.Errors)
	return CorrectionSummary{
		HasErrors:  errorCount > 0,
		ErrorCount: errorCount,
	}
}

// scoreSummaryFromResult 从评分结果生成摘要。
func scoreSummaryFromResult(score model.ScoreResult) ScoreSummary {
	return ScoreSummary{
		TotalScore: score.TotalScore,
		Grammar:    score.Grammar,
		Expression: score.Expression,
	}
}

// saveMessageState 保存消息快照到短期状态存储。
func (s *SessionService) saveMessageState(ctx context.Context, session model.Session, stage string) error {
	if s.stateStore == nil {
		return nil
	}
	if err := s.stateStore.SaveMessageSnapshot(ctx, session.ID, session.Messages); err != nil {
		return fmt.Errorf("%w: 保存消息快照失败：%v", ErrStateStoreFailed, err)
	}
	return s.saveSessionState(ctx, session, stage)
}

// saveSessionState 保存 Session 快照到短期状态存储。
func (s *SessionService) saveSessionState(ctx context.Context, session model.Session, stage string) error {
	if s.stateStore == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if stage == "" {
		stage = currentSessionStage(session)
	}
	err := s.stateStore.SaveSessionState(ctx, state.SessionState{
		SessionID:  session.ID,
		ScenarioID: session.ScenarioID,
		UserID:     session.UserID,
		Status:     string(session.Status),
		Stage:      stage,
		TurnCount:  session.TurnCount,
		UpdatedAt:  s.now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("%w: 保存 Session 状态失败：%v", ErrStateStoreFailed, err)
	}

	return nil
}

// appendCorrectionState 将纠错结果追加到短期状态存储。
func (s *SessionService) appendCorrectionState(ctx context.Context, correction model.CorrectionResult) error {
	if s.stateStore == nil {
		return nil
	}
	if err := s.stateStore.AppendCorrection(ctx, correction); err != nil {
		return fmt.Errorf("%w: 保存纠错状态失败：%v", ErrStateStoreFailed, err)
	}

	return nil
}

// savePartialScoreState 保存当前评分到短期状态存储。
func (s *SessionService) savePartialScoreState(ctx context.Context, score model.ScoreResult) error {
	if s.stateStore == nil {
		return nil
	}
	if err := s.stateStore.SavePartialScore(ctx, score); err != nil {
		return fmt.Errorf("%w: 保存当前评分失败：%v", ErrStateStoreFailed, err)
	}

	return nil
}

// currentSessionStage 返回 Session 当前训练阶段。
func currentSessionStage(session model.Session) string {
	if len(session.Messages) == 0 {
		return ""
	}

	return session.Messages[len(session.Messages)-1].Stage
}

// publishStreamEvent 发布训练流式业务事件。
func (s *SessionService) publishStreamEvent(event stream.Event) error {
	if s.events == nil {
		return nil
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = s.now()
	}
	if err := s.events.Publish(event); err != nil {
		return fmt.Errorf("%w: %v", ErrEventPublishFailed, err)
	}

	return nil
}

// publishSessionError 发布训练消息链路错误事件。
func (s *SessionService) publishSessionError(sessionID int, err error) {
	if s.events == nil {
		return
	}

	code, message := sessionErrorPayload(err)
	s.publishStreamEvent(stream.Event{
		Type:      stream.EventTypeError,
		SessionID: sessionID,
		Payload: stream.ErrorPayload{
			Code:    code,
			Message: message,
		},
	})
}

// sessionErrorPayload 将 Session 错误转换为 SSE 错误载荷。
func sessionErrorPayload(err error) (string, string) {
	switch {
	case errors.Is(err, ErrConversationAgentFailed):
		return "conversation_agent_failed", "对话 AI 回复失败"
	case errors.Is(err, ErrFeedbackAgentFailed):
		return "feedback_agent_failed", "反馈 AI 生成失败"
	case errors.Is(err, ErrSessionNotFound):
		return "session_not_found", "未找到训练"
	case errors.Is(err, ErrSessionAlreadyFinished):
		return "session_already_finished", "训练已结束"
	default:
		return "message_send_failed", "消息发送失败"
	}
}
