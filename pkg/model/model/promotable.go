package model

type Promotable interface {
	GraphModel
	GetPendingPromotion() string
}

type PromotableEmbed struct {
	pendingPromotion string
}

func (p *PromotableEmbed) GetPendingPromotion() string {
	return p.pendingPromotion
}

const NO_PENDING_PROMOTION = ""

func PendingPromotion(model GraphModel) (string, bool) {
	if promotable, ok := model.(Promotable); ok {
		pendingPromotion := promotable.GetPendingPromotion()
		return pendingPromotion, pendingPromotion != NO_PENDING_PROMOTION
	}
	return NO_PENDING_PROMOTION, false
}

func IsPromotable(model GraphModel) bool {
	_, ok := model.(Promotable)
	return ok
}
