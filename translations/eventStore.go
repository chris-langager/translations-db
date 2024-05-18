package translations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
)

type EventStore interface {
	Write(ctx context.Context, event Event) error
	NewGenerator(queryOptions ...QueryOption) GeneratorFn
}

type InMemoryEventStore struct {
	events []Event
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events: []Event{
			ProjectCreated{
				EventBase: NewEventBase(context.Background(), "asdf"),
				Id:        "asdf",
				Name:      "Test Project",
			},
			KeyCreated{
				EventBase: NewEventBase(context.Background(), "asdf"),
				ProjectId: "asdf",
				Id:        "header_1",
			},
			TranslationUpdated{
				EventBase: NewEventBase(context.Background(), "asdf"),
				ProjectId: "asdf",
				KeyId:     "header_1",
				Id:        "en",
				Value:     "Hello",
			},
			TranslationUpdated{
				EventBase: NewEventBase(context.Background(), "asdf"),
				ProjectId: "asdf",
				KeyId:     "header_1",
				Id:        "es",
				Value:     "Hola",
			},
		},
	}
}

func (o *InMemoryEventStore) Handle(ctx context.Context, tx *sql.Tx, event Event) error {
	return o.Write(ctx, event)
}

func (o *InMemoryEventStore) Write(ctx context.Context, event Event) error {
	o.events = append(o.events, event)
	return nil
}

func (o *InMemoryEventStore) NewGenerator(queryOptions ...QueryOption) GeneratorFn {
	var query Query
	for _, opt := range queryOptions {
		opt(&query)
	}

	current := 0
	return func(ctx context.Context) (Event, error) {
		for i := current; i < len(o.events); i++ {
			current++
			if query.AggregateIds != nil && !Contains(query.AggregateIds, o.events[i].GetAggregateId()) {
				continue
			}

			return o.events[i], nil
		}
		return nil, nil
	}
}

func Contains[T comparable](tt []T, t T) bool {
	for _, elem := range tt {
		if elem == t {
			return true
		}
	}
	return false
}

type Query struct {
	AggregateIds []string
	Types        []string
}

type QueryOption func(query *Query)

func AggregateIds(aggregateIds ...string) QueryOption {
	return func(query *Query) {
		query.AggregateIds = aggregateIds
	}
}

type TypeWrapper struct {
	TypeName string `json:"typeName"`
	Payload  string `json:"payload"`
}

func Serialize(data any) (string, error) {
	serializedPayload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	serializedTypeWrapper, err := json.Marshal(TypeWrapper{
		TypeName: reflect.TypeOf(data).String(),
		Payload:  string(serializedPayload),
	})
	if err != nil {
		return "", err
	}
	return string(serializedTypeWrapper), nil
}

var ErrorNoTypeMatch = errors.New("no type match found in targets")

func Deserialize(data string, targets ...any) (any, error) {
	var typeWrapper TypeWrapper
	err := json.Unmarshal([]byte(data), &typeWrapper)
	if err != nil {
		return nil, err
	}

	for _, target := range targets {
		rTargetType := reflect.TypeOf(target)
		if typeWrapper.TypeName != rTargetType.String() {
			continue
		}

		rTargetPtr := reflect.New(rTargetType)                                     // this is like t := &T{}
		err := json.Unmarshal([]byte(typeWrapper.Payload), rTargetPtr.Interface()) // this is like t
		if err != nil {
			return nil, err
		}

		return rTargetPtr.Elem().Interface(), nil // this is like *t
	}

	return nil, ErrorNoTypeMatch
}
