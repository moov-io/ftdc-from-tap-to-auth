# Templates for building acquirer tasks

## Imports, if you don't have auto-imports configured

```go
import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/moov-io/iso8583"
	"github.com/moov-io/iso8583/encoding"
	"github.com/moov-io/iso8583/field"
	"github.com/moov-io/iso8583/padding"
	"github.com/moov-io/iso8583/prefix"
	"github.com/moov-io/iso8583/sort"
)
```

## ISO 8583 Specification Template

```
var spec *iso8583.MessageSpec = &iso8583.MessageSpec{
	Name: "FTDC ISO 8583 Specification",
	Fields: map[int]field.Field{
		0: field.NewString(&field.Spec{
			Length:      4,
			Description: "Message Type Indicator",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		1: field.NewBitmap(&field.Spec{
			Length:      8,
			Description: "Bitmap",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.Binary.Fixed,
		}),
        // TODO: add field 2 and 3 from the spec
		4: field.NewString(&field.Spec{
			Length:      20,
			Description: "Processing date time",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		5: field.NewString(&field.Spec{
			Length:      2,
			Description: "Approval Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		6: field.NewString(&field.Spec{
			Length:      6,
			Description: "Authorization Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		7: field.NewString(&field.Spec{
			Length:      3,
			Description: "Currency Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		8: field.NewString(&field.Spec{
			Length:      4,
			Description: "Card Verification Value",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		9: field.NewString(&field.Spec{
			Length:      4,
			Description: "Expiration Date",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		10: field.NewComposite(&field.Spec{
			Length:      999,
			Description: "Acceptor Information",
			Pref:        prefix.ASCII.LLL,
			Tag: &field.TagSpec{
				Length: 2,
				Enc:    encoding.ASCII,
				Sort:   sort.StringsByInt,
			},
			Subfields: map[string]field.Field{
                // TODO: add subfield 01 from the spec
				"02": field.NewString(&field.Spec{
					Length:      4,
					Description: "MCC",
					Enc:         encoding.ASCII,
					Pref:        prefix.ASCII.Fixed,
				}),
				"03": field.NewString(&field.Spec{
					Length:      10,
					Description: "Postal Code",
					Enc:         encoding.ASCII,
					Pref:        prefix.ASCII.LL,
				}),
				"04": field.NewString(&field.Spec{
					Length:      299,
					Description: "Merchant URL",
					Enc:         encoding.ASCII,
					Pref:        prefix.ASCII.LLL,
				}),
			},
		}),
		11: field.NewString(&field.Spec{
			Length:      6,
			Description: "System Trace Audit Number (STAN)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
	},
}
```

## Request and Response Data Types

```go
type AuthorizationRequest struct {
	MTI                   string        `iso8583:"0"`
    // TODO: define PAN and Amount fields
	ProcessingDateTime    string        `iso8583:"4"`
	CurrencyCode          string        `iso8583:"7"`
	CardVerificationValue string        `iso8583:"8"`
	ExpirationDate        string        `iso8583:"9"`
	AcceptorInfo          *AcceptorInfo `iso8583:"10"`
	STAN                  string        `iso8583:"11"`
}

type AcceptorInfo struct {
    // TODO: define field for tag 01
	MCC          string `iso8583:"02"`
	PostalCode   string `iso8583:"03"`
	MerchantURL  string `iso8583:"04"`
}

// 3. Define the AuthorizationResponse struct that matches the ISO 8583 specification
type AuthorizationResponse struct {
	MTI               string `iso8583:"0"`
	ApprovalCode      string `iso8583:"5"`
	AuthorizationCode string `iso8583:"6"`
}
```

## Populating Request Data

```go
	authRequest := AuthorizationRequest{
		MTI:                   "0100",
        // TODO: populate PAN and Amount fields
		ProcessingDateTime:    time.Now().UTC().Format(time.RFC3339),
		CurrencyCode:          "840",
		CardVerificationValue: "1234",
		ExpirationDate:        "2509",
		AcceptorInfo: &AcceptorInfo{
			MerchantName: "Fintech DevCon Demo",
			MCC:          "1234",
			PostalCode:   "12345",
			MerchantURL:  "https://fintechdevcon.com",
		},
		STAN: "123456",
	}
```


## Printing ISO 8583 Messages

```go
iso8583.Describe(unpackedMsg, os.Stdout)
```

## Writing message length header

```go
	// write message length as a 2-bytes header
	messageLength := len(packed)
	packedLength := make([]byte, 2)
	packedLength[0] = byte(messageLength >> 8) // high byte
	packedLength[1] = byte(messageLength)      // low byte

	_, err = conn.Write(packedLength)
	if err != nil {
		fmt.Printf("Error sending message length to FTDC Issuer: %v\n", err)
		return
	}
```

## Reading length header

```go
// read message length from the first 2 bytes
// and convert it to an integer
responseMessageLength := int(responseLength[0])<<8 | int(responseLength[1])
responsePacked := make([]byte, responseMessageLength)
```

