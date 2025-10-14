package model

type LabelSettable interface {
	GraphModel
	GetPendingLabelAddition() string
}

type LabelSettableEmbed struct {
	PendingLabelAddition string `json:"-" neo4j:"-" dynamodbav:"-" desc:"The label that is pending addition to the model in Neo4j." example:"Seed"`
}

func (p *LabelSettableEmbed) GetPendingLabelAddition() string {
	return p.PendingLabelAddition
}

const NO_PENDING_LABEL_ADDITION = ""

func PendingLabelAddition(model GraphModel) (string, bool) {
	if relabelable, ok := model.(LabelSettable); ok {
		pendingAddition := relabelable.GetPendingLabelAddition()
		return pendingAddition, pendingAddition != NO_PENDING_LABEL_ADDITION
	}
	return NO_PENDING_LABEL_ADDITION, false
}

func IsLabelSettable(model GraphModel) bool {
	_, ok := model.(LabelSettable)
	return ok
}
