package adapters

import (
	"clean-arch-go/app/query"
	"clean-arch-go/domain/errors"
	"clean-arch-go/domain/task"
	"context"
	"fmt"
)

type TaskUpdater func(context.Context, task.Task) (task.Task, error)

type TaskPgRepository struct {
}

func NewTaskPgRepository() TaskPgRepository {
	return TaskPgRepository{}
}

func (r TaskPgRepository) Add(ctx context.Context, t task.Task) error {
	return nil
}

func (r TaskPgRepository) UpdateByID(ctx context.Context, uuid string, updateFn TaskUpdater) error {
	return nil
}

func (r TaskPgRepository) FindTasks(ctx context.Context) ([]query.Task, error) {
	return []query.Task{}, nil
}

func (r TaskPgRepository) FindById(ctx context.Context, uuid string) (task.Task, error) {
	return nil, nil
}

func (r TaskPgRepository) FindTasksForUser(ctx context.Context, userUUID string) ([]query.Task, error) {
	return []query.Task{}, errors.E(errors.Op("task.FindTasksForUser"), errors.Internal, fmt.Errorf("dump"))
}
