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

// IsSeedPromotion returns true if other represents a seed promotion for current.
func IsSeedPromotion(current, other *BaseAsset) bool {
	return current.Source != SeedSource && other.Source == SeedSource
}

// ApplySeedLabels sets the seed label and source on a model being promoted.
// Use this in Visit paths where no history record should be created.
func ApplySeedLabels(base *BaseAsset, ls *LabelSettableEmbed) {
	ls.PendingLabelAddition = SeedLabel
	base.Source = SeedSource
}

// PromoteToSeed handles full seed promotion during Merge: sets labels, source,
// and records a promotion history event. The empty From with non-empty To
// signals a promotion event to the UI.
func PromoteToSeed(base *BaseAsset, ls *LabelSettableEmbed, targetStatus string) {
	ApplySeedLabels(base, ls)
	base.History.RecordPromotion("", targetStatus)
}

// MergeWithPromotionCheck dispatches to the correct merge path for models
// embedding both BaseAsset and LabelSettableEmbed. Seed promotions use
// PromoteToSeed + MergeFields; standard updates use BaseAsset.Merge.
func MergeWithPromotionCheck(base *BaseAsset, ls *LabelSettableEmbed, other Assetlike) {
	otherBase := other.GetBase()
	if IsSeedPromotion(base, otherBase) {
		targetStatus := otherBase.Status
		if targetStatus == "" {
			targetStatus = base.Status
		}
		PromoteToSeed(base, ls, targetStatus)
		if targetStatus != base.Status {
			base.Status = targetStatus
		}
		base.MergeFields(other)
	} else {
		base.Merge(other)
	}
}
