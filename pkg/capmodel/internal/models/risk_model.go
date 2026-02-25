package models

// Risk is the capmodel representation of a security finding. Target accepts
// any capmodel type so the emitter can resolve the correct chariot model type
// at runtime (Asset, Repository, etc.).
type Risk struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Status string `json:"status"`
	Proof  []byte `json:"proof"`
	Target any    `json:"target,omitempty"`
}
