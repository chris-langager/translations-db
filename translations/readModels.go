package translations

import (
	"context"
	"database/sql"
	"sync"
)

type InMemoryProjectList struct {
	projectsById map[string]*Project
}

func NewInMemoryProjectList() *InMemoryProjectList {
	return &InMemoryProjectList{
		projectsById: map[string]*Project{},
	}
}

func (o *InMemoryProjectList) GetProject(id string) *Project {
	return o.projectsById[id]
}

func (o *InMemoryProjectList) ListProjects() []Project {
	ret := []Project{}
	for _, project := range o.projectsById {
		ret = append(ret, *project)
	}
	return ret
}

func (o *InMemoryProjectList) Handle(ctx context.Context, tx *sql.Tx, event Event) error {
	sync.OnceFunc(func() {
		o.projectsById = map[string]*Project{}
	})()

	if o.projectsById[event.GetAggregateId()] == nil {
		o.projectsById[event.GetAggregateId()] = &Project{}
	}

	o.projectsById[event.GetAggregateId()].Reduce(event)
	return nil
}
