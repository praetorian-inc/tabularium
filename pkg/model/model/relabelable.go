package model

type Relabelable interface {
	GraphModel
	GetPendingLabelAddition() string
}

type RelabelableEmbed struct {
	PendingLabelAddition string `json:"-" neo4j:"-" dynamodbav:"-" desc:"The label that is pending addition to the model in Neo4j." example:"Seed"`
}

func (p *RelabelableEmbed) GetPendingLabelAddition() string {
	return p.PendingLabelAddition
}

const NO_PENDING_LABEL_ADDITION = ""

func PendingLabelAddition(model GraphModel) (string, bool) {
	if relabelable, ok := model.(Relabelable); ok {
		pendingAddition := relabelable.GetPendingLabelAddition()
		return pendingAddition, pendingAddition != NO_PENDING_LABEL_ADDITION
	}
	return NO_PENDING_LABEL_ADDITION, false
}

func IsRelabelable(model GraphModel) bool {
	_, ok := model.(Relabelable)
	return ok
}
