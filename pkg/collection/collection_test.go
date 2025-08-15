package collection

import (
	"reflect"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/lib/plural"
	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/stretchr/testify/assert"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// testModelA is a simple implementation of registry.Model for testing.
type testModelA struct {
	registry.BaseModel
	ID   int
	Name string
}

func (tm *testModelA) GetDescription() string {
	return "This is test model A"
}

// testModelB is another simple implementation of registry.Model for testing.
type testModelB struct {
	registry.BaseModel
	Value float64
}

func (tm *testModelB) GetDescription() string {
	return "This is test model B"
}

// testModelC is an empty struct implementing registry.Model for testing edge cases.
type testModelC struct {
	registry.BaseModel
}

func (tm *testModelC) GetDescription() string {
	return "This is test model C"
}

// unregisteredModel is a model type used for testing Get's behavior with types
// not fully present in the registry during a Get operation.
// It must still implement registry.Model to be added to a collection.
type unregisteredModel struct {
	registry.BaseModel
	Data string
}

func (um *unregisteredModel) GetDescription() string { return "Unregistered" }

// getTestAuxiliaryModel is a model type used for testing Get behavior, specifically
// for scenarios where a registered type is queried but not present in the collection.
// It implements registry.Model.
type getTestAuxiliaryModel struct {
	registry.BaseModel
	DummyField int
}

func (gtam *getTestAuxiliaryModel) GetDescription() string {
	return "Auxiliary model for Get testing (e.g., not expected in collection)."
}

// testInterface is an interface that testModelA implements.
type testInterface interface {
	registry.Model
	GetName() string
}

func (tm *testModelA) GetName() string {
	return tm.Name
}

func init() {
	// Ensure a clean registry for this test package
	registry.Registry = registry.NewTypeRegistry()

	// Register all test models
	registry.Registry.MustRegisterModel(&testModelA{})
	registry.Registry.MustRegisterModel(&testModelB{})
	registry.Registry.MustRegisterModel(&testModelC{})
	registry.Registry.MustRegisterModel(&unregisteredModel{})
	registry.Registry.MustRegisterModel(&getTestAuxiliaryModel{})
}

func TestNewCollection(t *testing.T) {
	c := Collection{}
	if len(c.Items) != 0 {
		t.Errorf("Collection{}.Items should be empty, got %v", c.Items)
	}
	if c.Count != 0 {
		t.Errorf("Collection{}.Count should be 0, got %d", c.Count)
	}
}

func TestCollection_AddAndCount(t *testing.T) {
	c := Collection{}

	modelA1 := &testModelA{ID: 1, Name: "A1"}
	modelA2 := &testModelA{ID: 2, Name: "A2"}
	modelB1 := &testModelB{Value: 1.1}

	c.Add(modelA1)
	if c.Count != 1 {
		t.Errorf("Count after adding one item, got %d, want %d", c.Count, 1)
	}
	nameA := plural.Plural(registry.Name(modelA1))
	if len(c.Items[nameA]) != 1 || !reflect.DeepEqual(c.Items[nameA][0], modelA1) {
		t.Errorf("Add() did not add modelA1 correctly. Items: %v", c.Items)
	}

	c.Add(modelA2)
	if c.Count != 2 {
		t.Errorf("Count after adding second item of same type, got %d, want %d", c.Count, 2)
	}
	if len(c.Items[nameA]) != 2 || !reflect.DeepEqual(c.Items[nameA][1], modelA2) {
		t.Errorf("Add() did not add modelA2 correctly. Items: %v", c.Items)
	}

	c.Add(modelB1)
	if c.Count != 3 {
		t.Errorf("Count after adding item of different type, got %d, want %d", c.Count, 3)
	}
	nameB := plural.Plural(registry.Name(modelB1))
	if len(c.Items[nameB]) != 1 || !reflect.DeepEqual(c.Items[nameB][0], modelB1) {
		t.Errorf("Add() did not add modelB1 correctly. Items: %v", c.Items)
	}
}

func TestGet(t *testing.T) {
	c := Collection{}

	modelA1 := &testModelA{ID: 1, Name: "A1"}
	modelA2 := &testModelA{ID: 2, Name: "A2"}
	modelB1 := &testModelB{Value: 1.1}
	modelC1 := &testModelC{}

	c.Add(modelA1)
	c.Add(modelB1)
	c.Add(modelA2)
	c.Add(modelC1)

	t.Run("Get specific type testModelA", func(t *testing.T) {
		resultsA := Get[*testModelA](&c)
		if len(resultsA) != 2 {
			t.Errorf("Get[*testModelA]() returned %d Items, want %d. Got: %v", len(resultsA), 2, resultsA)
		}
		// Check if we got the correct Items (order might not be guaranteed by Add)
		foundA1, foundA2 := false, false
		for _, item := range resultsA {
			if reflect.DeepEqual(item, modelA1) {
				foundA1 = true
			}
			if reflect.DeepEqual(item, modelA2) {
				foundA2 = true
			}
		}
		if !foundA1 || !foundA2 {
			t.Errorf("Get[*testModelA]() did not return the correct Items. Expected: [%v, %v], Got: %v", modelA1, modelA2, resultsA)
		}
	})

	t.Run("Get specific type testModelB", func(t *testing.T) {
		resultsB := Get[*testModelB](&c)
		if len(resultsB) != 1 {
			t.Errorf("Get[*testModelB]() returned %d Items, want %d", len(resultsB), 1)
		}
		if !reflect.DeepEqual(resultsB[0], modelB1) {
			t.Errorf("Get[*testModelB]() did not return the correct item. Expected: %v, Got: %v", modelB1, resultsB[0])
		}
	})

	t.Run("Get specific type testModelC", func(t *testing.T) {
		resultsC := Get[*testModelC](&c)
		if len(resultsC) != 1 {
			t.Errorf("Get[*testModelC]() returned %d Items, want %d", len(resultsC), 1)
		}
		if !reflect.DeepEqual(resultsC[0], modelC1) {
			t.Errorf("Get[*testModelC]() did not return the correct item. Expected: %v, Got: %v", modelC1, resultsC[0])
		}
	})

	t.Run("Get type not in collection", func(t *testing.T) {
		// Ensure getTestAuxiliaryModel is registered for this sub-test if not already globally.
		// This is now handled by init()
		// registry.Registry.MustRegisterModel(&getTestAuxiliaryModel{})

		resultsNP := Get[*getTestAuxiliaryModel](&c)
		if len(resultsNP) != 0 {
			t.Errorf("Get[*getTestAuxiliaryModel]() returned %d Items, want %d", len(resultsNP), 0)
		}
	})

	t.Run("Get from empty collection", func(t *testing.T) {
		emptyC := Collection{}
		resultsA := Get[*testModelA](&emptyC)
		if len(resultsA) != 0 {
			t.Errorf("Get[*testModelA]() from empty collection returned %d Items, want %d", len(resultsA), 0)
		}
	})

	t.Run("Get by interface", func(t *testing.T) {
		isolatedC := Collection{}

		mA1 := &testModelA{ID: 10, Name: "InterfaceA1"}
		mA2 := &testModelA{ID: 20, Name: "InterfaceA2"}
		mB1 := &testModelB{Value: 10.1} // Does not implement testInterface

		isolatedC.Add(mA1)
		isolatedC.Add(mB1) // Add a non-matching type
		isolatedC.Add(mA2)

		resultsI := Get[testInterface](&isolatedC)
		if len(resultsI) != 2 {
			t.Errorf("Get[testInterface]() returned %d Items, want %d. Got: %v", len(resultsI), 2, resultsI)
		}

		foundIA1, foundIA2 := false, false
		for _, item := range resultsI {
			if item.GetName() == "InterfaceA1" { // Use interface method
				foundIA1 = true
			}
			if item.GetName() == "InterfaceA2" {
				foundIA2 = true
			}
		}
		if !foundIA1 || !foundIA2 {
			t.Errorf("Get[testInterface]() did not return the correct Items. Expected names: [InterfaceA1, InterfaceA2], Got: %v", resultsI)
		}
	})
}

// TestGet_WithUnregisteredTypeInCollection (Illustrative - Get relies on registry for type checking)
func TestGet_WithUnregisteredTypeInCollection(t *testing.T) {
	c := Collection{}

	itemUnregistered := &unregisteredModel{Data: "secret"}
	itemA := &testModelA{ID: 100, Name: "GetterA"}
	c.Add(itemUnregistered)
	c.Add(itemA)

	// Attempt to get testModelA Items
	resultsA := Get[*testModelA](&c)
	if len(resultsA) != 1 {
		t.Errorf("Get[*testModelA]() when collection also has other types, got %d Items, want 1. Items: %v", len(resultsA), resultsA)
	} else if !reflect.DeepEqual(resultsA[0], itemA) {
		t.Errorf("Get[*testModelA]() returned wrong item. Got: %v, Want: %v", resultsA[0], itemA)
	}

	// Attempt to get unregisteredModel Items. Since unregisteredModel is registered in init(), it should be found.
	resultsUnregistered := Get[*unregisteredModel](&c)
	if len(resultsUnregistered) != 1 {
		t.Errorf("Get[*unregisteredModel]() returned %d Items, want 1 (should be found as it is registered in init). Items: %v", len(resultsUnregistered), resultsUnregistered)
	} else if !reflect.DeepEqual(resultsUnregistered[0], itemUnregistered) {
		t.Errorf("Get[*unregisteredModel]() returned wrong item. Got: %v, Want: %v", resultsUnregistered[0], itemUnregistered)
	}
}

func TestAddInterface_Seed(t *testing.T) {
	c := Collection{Label: model.SeedLabel}

	seed1 := model.NewAssetSeed("example.com")
	seed2 := model.NewAssetSeed("1.2.3.4")
	seed3 := model.NewAssetSeed("1.2.3.4/24")
	asset := model.NewAsset("example.com", "example.com")
	attribute := model.NewAttribute("name", "value", &asset)

	c.Add(&seed1)
	c.Add(&seed2)
	c.Add(&seed3)
	c.Add(&asset)
	c.Add(&attribute)

	assert.Len(t, c.Items, 3)
	assert.Len(t, c.Items["seeds"], 3)
	assert.Len(t, c.Items["assets"], 1)
	assert.Len(t, c.Items["attributes"], 1)
}
