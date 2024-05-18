package translations

import (
	"context"
	"encoding/json"
	"time"
)

/*
Aggregates
- any kind of state that implements a Reduce function
*/

type Generator interface {
	Next(ctx context.Context) (Event, error)
}
type GeneratorFn func(ctx context.Context) (Event, error)

func (fn GeneratorFn) Next(ctx context.Context) (Event, error) {
	return fn(ctx)
}

func ReduceWith[T Aggregate](ctx context.Context, t T, generator Generator) error {
	for {
		event, err := generator.Next(ctx)
		if err != nil {
			return err
		}
		if event == nil {
			break
		}
		t.Reduce(event)
	}
	return nil
}

type Aggregate interface {
	Reduce(event Event)
}

type Project struct {
	Id          string
	Name        string
	DateCreated time.Time
	DateUpdated time.Time
	Locales     []string
	KeysById    map[string]*Key

	History []string
}

type Key struct {
	Id          string
	DateCreated time.Time
	DateUpdated time.Time

	TranslationsById map[string]*Translation
}

type Translation struct {
	Id          string
	DateCreated time.Time
	DateUpdated time.Time

	Value string

	ProjectId string
	KeyId     string
}

func (o *Project) Reduce(event Event) {
	switch e := event.(type) {
	case ProjectCreated:
		*o = Project{
			Id:          e.Id,
			Name:        e.Name,
			DateCreated: e.Timestamp,
			DateUpdated: e.Timestamp,
			Locales:     []string{"es", "en"},
			KeysById:    map[string]*Key{},
		}
	case ProjectUpdated:
		o.Name = e.Name
		o.DateUpdated = e.Timestamp
	case ProjectDeleted:
		o = nil
	case KeyCreated:
		key := &Key{
			Id:               e.Id,
			DateCreated:      e.Timestamp,
			DateUpdated:      e.Timestamp,
			TranslationsById: map[string]*Translation{},
		}
		for _, locale := range o.Locales {
			key.TranslationsById[locale] = &Translation{
				ProjectId:   e.ProjectId,
				KeyId:       e.Id,
				Id:          locale,
				DateCreated: e.Timestamp,
				DateUpdated: e.Timestamp,
			}
		}
		o.KeysById[e.Id] = key
	case KeyDeleted:
		delete(o.KeysById, e.Id)
	// case TranslationCreated:
	// 	key, ok := o.KeysById[e.KeyId]
	// 	if !ok {
	// 		break
	// 	}
	// 	key.DateUpdated = e.Timestamp
	// 	key.TranslationsById[e.Id] = &Translation{
	// 		Id:          e.Id,
	// 		DateCreated: e.Timestamp,
	// 		DateUpdated: e.Timestamp,
	// 		Value:       e.Value,
	// 	}
	case TranslationUpdated:
		key, ok := o.KeysById[e.KeyId]
		if !ok {
			break
		}
		key.DateUpdated = e.Timestamp
		key.TranslationsById[e.Id].DateUpdated = e.Timestamp
		key.TranslationsById[e.Id].Value = e.Value
	case TranslationDeleted:
		key, ok := o.KeysById[e.KeyId]
		if !ok {
			break
		}
		delete(key.TranslationsById, e.Id)
	}

	o.DateUpdated = event.GetTimestamp()

	h, _ := json.MarshalIndent(event, "", "  ")
	o.History = append([]string{string(h)}, o.History...)
}

type ProjectList struct {
	ProjectsById map[string]*Project
	History      []string
}

func (o *ProjectList) Reduce(event Event) {
	if o.ProjectsById == nil {
		o.ProjectsById = map[string]*Project{}
	}

	switch e := event.(type) {
	case ProjectCreated:
		project := &Project{}
		project.Reduce(event)
		o.ProjectsById[e.Id] = project
	case ProjectUpdated:
		o.ProjectsById[e.Id].Reduce(event)
	case ProjectDeleted:
		delete(o.ProjectsById, e.Id)
	}

	s, _ := json.MarshalIndent(event, "", "  ")
	o.History = append(o.History, string(s))
}
