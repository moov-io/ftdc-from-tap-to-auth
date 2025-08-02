# Receipt Printer Service

This service is designed to print receipts for payment transactions.

## Start

To start the service, run the following command (from the project root):

```bash
go run ./cmd/printer
```

## Usage

```
curl -X POST \
        -H "Content-Type: application/json" \
        -d '{
      "payment_id": "txn_abc123def456",
      "processed_at": "2025-08-02T14:30:45Z",
      "pan": "4532********1234",
      "cardholder": "JOHN DOE",
      "amount": 2599,
      "authorization_code": "123456",
      "response_code": "00"
    }' http://0.0.0.0:8085/receipts
```
