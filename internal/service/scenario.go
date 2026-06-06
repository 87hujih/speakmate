package service

import (
	"errors"

	"speakmate/internal/model"
	"speakmate/internal/repository"
)

// ErrScenarioNotFound 表示业务层没有找到对应场景。
var ErrScenarioNotFound = errors.New("scenario not found")

// ScenarioRepository 定义场景服务依赖的数据访问能力。
type ScenarioRepository interface {
	List() []model.Scenario
	FindByID(id int) (model.Scenario, error)
}

// ScenarioService 封装场景查询的业务流程和错误语义。
type ScenarioService struct {
	repo ScenarioRepository
}

// NewScenarioService 创建场景服务实例。
func NewScenarioService(repo ScenarioRepository) *ScenarioService {
	return &ScenarioService{
		repo: repo,
	}
}

// ListScenarios 返回所有可训练场景。
func (s *ScenarioService) ListScenarios() []model.Scenario {
	return s.repo.List()
}

// GetScenario 查询单个场景，并把仓库错误转换成业务层错误。
func (s *ScenarioService) GetScenario(id int) (model.Scenario, error) {
	scenario, err := s.repo.FindByID(id)
	if err == nil {
		return scenario, nil
	}
	if errors.Is(err, repository.ErrScenarioNotFound) {
		return model.Scenario{}, ErrScenarioNotFound
	}

	// 非预期仓库错误原样向上返回，交给 HTTP 层转换成 500。
	return model.Scenario{}, err
}
