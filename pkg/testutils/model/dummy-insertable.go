package model

import "github.com/praetorian-inc/tabularium/pkg/registry"

func init() {
	registry.Registry.MustRegisterModel(&DummyInsertable{})
}

type DummyInsertable struct {
	registry.BaseModel
	Merged  bool
	Merges  int
	Visited bool
	Visits  int
	Status  string
}

func NewDummyInsertable() *DummyInsertable {
	return &DummyInsertable{}
}

func (i *DummyInsertable) GetKey() string {
	return "#dummyinsertable"
}

func (i *DummyInsertable) GetDescription() string {
	return "insertable object used for unit tests"
}

func (i *DummyInsertable) GetLabels() []string {
	return []string{"dummyinsertable"}
}

func (i *DummyInsertable) SetStatus(status string) {
	i.Status = status
}

func (i *DummyInsertable) Merge(item any) {
	di, ok := item.(*DummyInsertable)
	if !ok {
		return
	}
	i.Merges += di.Merges
	i.Visits += di.Visits
}

func (i *DummyInsertable) Visit(_ any) error {
	i.Visited = true
	i.Visits++
	return nil
}

func (i *DummyInsertable) Valid() bool {
	return true
}
