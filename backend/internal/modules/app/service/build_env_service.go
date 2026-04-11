package service

import (
	"errors"
	"sync"
	"time"

	"devops-platform/internal/modules/app/model"
)

type BuildEnvService struct {
	mu      sync.RWMutex
	envs    map[uint]model.BuildEnv
	nextID  uint
}

func NewBuildEnvService() *BuildEnvService {
	return &BuildEnvService{
		envs:   make(map[uint]model.BuildEnv),
		nextID: 1,
	}
}

func (s *BuildEnvService) CreateBuildEnv(env model.BuildEnv) (model.BuildEnv, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	env.ID = s.nextID
	s.nextID++
	env.CreatedAt = time.Now()
	env.UpdatedAt = time.Now()
	s.envs[env.ID] = env
	return env, nil
}

func (s *BuildEnvService) ListBuildEnvs() []model.BuildEnv {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]model.BuildEnv, 0, len(s.envs))
	for _, env := range s.envs {
		list = append(list, env)
	}
	return list
}

func (s *BuildEnvService) UpdateBuildEnv(env model.BuildEnv) (model.BuildEnv, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.envs[env.ID]; !ok {
		return model.BuildEnv{}, errors.New("构建环境不存在")
	}
	env.UpdatedAt = time.Now()
	s.envs[env.ID] = env
	return env, nil
}

func (s *BuildEnvService) DeleteBuildEnv(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.envs[id]; !ok {
		return errors.New("构建环境不存在")
	}
	delete(s.envs, id)
	return nil
}
