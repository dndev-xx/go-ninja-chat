package managerload

import (
	"context"
	"fmt"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (s *Service) CanManagerTakeProblem(ctx context.Context, managerID types.UserID) (bool, error) {
	count, err := s.problemsRepo.GetManagerOpenProblemsCount(ctx, managerID)
	if err != nil {
		return false, fmt.Errorf("get manager open problems count: %v", err)
	}
	if count < s.maxProblemsAtTime {
		return true, nil
	}
	return false, nil
}
