package main

import "time"

type Envelope[T any] struct {
	Id string

	Actor       string    // who
	Type        string    // what
	Timestamp   time.Time // when
	AggregateId string    // where
	Payload     T
}

type Event interface {
	GetType() string
	GetAggregateId() string
}

type Reducer[T any] interface {
	Reduce(t *T)
}

type ProjectCreatedPayload struct {
	Id   string
	Name string
}
type ProjectCreated Envelope[ProjectCreatedPayload]

func (o ProjectCreated) GetType() string {
	return "ProjectCreated"
}

func (o ProjectCreated) GetAggregateId() string {
	return o.AggregateId
}

type KeyCreatedPayload struct {
	Id        string
	ProjectId string
}
type KeyCreated Envelope[KeyCreatedPayload]

func (o KeyCreated) GetType() string {
	return "KeyCreated"
}
func (o KeyCreated) GetAggregateId() string {
	return o.AggregateId
}

type KeyDeletedPayload struct {
	Id        string
	ProjectId string
}
type KeyDeleted Envelope[KeyDeletedPayload]

func (o KeyDeleted) GetType() string {
	return "KeyDeleted"
}
func (o KeyDeleted) GetAggregateId() string {
	return o.AggregateId
}

// type TranslationCreatedPayload struct {
// 	Id        string
// 	ProjectId string
// 	KeyId     string

// 	Value string
// }
// type TranslationCreated Envelope[TranslationCreatedPayload]

// func (o TranslationCreated) GetType() string {
// 	return "TranslationCreated"
// }
// func (o TranslationCreated) GetAggregateId() string {
// 	return o.AggregateId
// }

type TranslationUpdatedPayload struct {
	Id        string
	ProjectId string
	KeyId     string

	Value string
}
type TranslationUpdated Envelope[TranslationUpdatedPayload]

func (o TranslationUpdated) GetType() string {
	return "TranslationUpdated"
}
func (o TranslationUpdated) GetAggregateId() string {
	return o.AggregateId
}

type TranslationDeletedPayload struct {
	Id        string
	ProjectId string
	KeyId     string
}
type TranslationDeleted Envelope[TranslationDeletedPayload]

func (o TranslationDeleted) GetType() string {
	return "TranslationDeleted"
}
func (o TranslationDeleted) GetAggregateId() string {
	return o.AggregateId
}

// Aggregates
type App struct {
	ProjectsById map[string]*Project
	History      []string
}

type Project struct {
	Id          string
	Name        string
	DateCreated time.Time
	DateUpdated time.Time

	KeysById map[string]*Key

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
}
