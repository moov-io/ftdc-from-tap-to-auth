package printer

import "time"

type Receipt struct {
	PaymentID          string    `json:"payment_id"`
	ProcessingDateTime time.Time `json:"processed_at"`
	PAN                string    `json:"pan"`
	Cardholder         string    `json:"cardholder"`
	Amount             int64     `json:"amount"`
	AuthorizationCode  string    `json:"authorization_code"`
	ResponseCode       string    `json:"response_code"`
}

// information about the printing job
type PrintJob struct {
	NumberInQueue int `json:"number_in_queue"`
	WaitingTime   int `json:"waiting_time"`
}
