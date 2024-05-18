package translations

import (
	"context"
	"time"
)

type Event interface {
	GetActor() string
	GetAggregateId() string
	GetTimestamp() time.Time
}

type EventBase struct {
	Actor       string    // who
	AggregateId string    // what (this aggregates into)
	Timestamp   time.Time // when
}

/*
Consider - should aggregateType be set?
*/
func NewEventBase(ctx context.Context, aggregateId string) EventBase {
	return EventBase{
		Actor:       GetActor(ctx),
		AggregateId: aggregateId,
		Timestamp:   time.Now(),
	}
}

func GetActor(ctx context.Context) string {
	//TODO
	return "123"
}

func (o EventBase) GetActor() string {
	return o.Actor
}

func (o EventBase) GetAggregateId() string {
	return o.AggregateId
}

func (o EventBase) GetTimestamp() time.Time {
	return o.Timestamp
}

/*
Events
- describes a thing that happened at a specific time
*/
type ProjectCreated struct {
	EventBase
	Id   string
	Name string
}

type ProjectUpdated struct {
	EventBase
	Id   string
	Name string
}

type ProjectDeleted struct {
	EventBase
	Id string
}

type KeyCreated struct {
	EventBase
	Id        string
	ProjectId string
}

type KeyDeleted struct {
	EventBase
	Id        string
	ProjectId string
}

// type TranslationCreated struct {
// 	EventBase
// 	Id        string
// 	KeyId     string
// 	ProjectId string
// 	Value     string
// }

type TranslationUpdated struct {
	EventBase
	Id        string
	KeyId     string
	ProjectId string
	Value     string
}

type TranslationDeleted struct {
	EventBase
	Id        string
	KeyId     string
	ProjectId string
}
