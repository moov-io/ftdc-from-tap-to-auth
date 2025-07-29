package models

type AuthorizationRequest struct {
	Amount     int64
	Currency   string
	Card       Card
	Merchant   Merchant
	EMVPayload []byte
}

type AuthorizationResponse struct {
	AuthorizationCode string
	ApprovalCode      string
}
