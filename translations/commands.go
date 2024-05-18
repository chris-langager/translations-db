package translations

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

/*
Commands
- take input, return event(s) or an error
- can read from whatever dependencies they want
*/

type Command[T any] func(context.Context, T) (Event, error)

type ReadModel interface {
	Handle(ctx context.Context, tx *sql.Tx, event Event) error
}

func NewCommandPipeline[T any](db *sql.DB, command Command[T], readModels ...ReadModel) func(context.Context, T) error {
	return func(ctx context.Context, t T) error {
		event, err := command(ctx, t)
		if err != nil {
			return err
		}

		tx, err := db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return err
		}
		for _, readModel := range readModels {
			err = readModel.Handle(ctx, tx, event)
			if err != nil {
				mustRollback(tx)
				return err
			}
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
		return nil
	}
}

func mustRollback(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		panic(err)
	}
}

type CreateProjectInput struct {
	Name string
}

func CreateProject() func(ctx context.Context, input CreateProjectInput) (Event, error) {
	return func(ctx context.Context, input CreateProjectInput) (Event, error) {
		id := uuid.NewString()
		return ProjectCreated{
			EventBase: NewEventBase(ctx, id),
			Id:        id,
			Name:      input.Name,
		}, nil
	}
}

type UpdateProjectInput struct {
	Id   string
	Name string
}

func UpdateProject(eventStore EventStore) func(ctx context.Context, input UpdateProjectInput) (Event, error) {
	return func(ctx context.Context, input UpdateProjectInput) (Event, error) {
		_, err := GetProject(ctx, eventStore, input.Id)
		if err != nil {
			return nil, err
		}

		return ProjectUpdated{
			EventBase: NewEventBase(ctx, input.Id),
			Id:        input.Id,
			Name:      input.Name,
		}, nil
	}
}

type CreateKeyInput struct {
	ProjectId string
	Id        string
}

func CreateKey() func(ctx context.Context, input CreateKeyInput) (Event, error) {
	return func(ctx context.Context, input CreateKeyInput) (Event, error) {
		return KeyCreated{
			EventBase: NewEventBase(ctx, input.ProjectId),
			ProjectId: input.ProjectId,
			Id:        input.Id,
		}, nil
	}
}

type UpdateTranslationInput struct {
	ProjectId string
	KeyId     string
	Id        string
	Value     string
}

func UpdateTranslation() func(ctx context.Context, input UpdateTranslationInput) (Event, error) {
	return func(ctx context.Context, input UpdateTranslationInput) (Event, error) {
		return TranslationUpdated{
			EventBase: NewEventBase(ctx, input.ProjectId),
			ProjectId: input.ProjectId,
			KeyId:     input.KeyId,
			Id:        input.Id,
			Value:     input.Value,
		}, nil
	}
}

var ErrorNotFound = errors.New("not found")

func GetProject(ctx context.Context, eventStore EventStore, id string) (*Project, error) {
	var project Project
	err := ReduceWith(ctx, &project, eventStore.NewGenerator(AggregateIds(id)))
	if project.Id == "" {
		return nil, ErrorNotFound
	}
	return &project, err

}

func GetProjectList(ctx context.Context, eventStore EventStore) (*ProjectList, error) {
	var projectList ProjectList
	err := ReduceWith(ctx, &projectList, eventStore.NewGenerator())
	return &projectList, err
}
