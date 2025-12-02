package model

// Hook represents a function that can be called as part of a model's lifecycle.
type Hook struct {
	Call        func() error
	Description string
}

// CallHooks executes all hooks associated with a model.
func CallHooks(model Model) error {
	for _, hook := range model.GetHooks() {
		if err := hook.Call(); err != nil {
			return err
		}
	}
	return nil
}
