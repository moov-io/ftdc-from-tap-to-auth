package main

import (
	"fmt"
	"net"
	"time"

	"github.com/moov-io/iso8583"
	"github.com/moov-io/iso8583/encoding"
	"github.com/moov-io/iso8583/field"
	"github.com/moov-io/iso8583/prefix"
	"github.com/moov-io/iso8583/sort"
)

// Step 1. Define the ISO 8583 message specification for FTDC Acquirer
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
		// TODO: add field 2 and field 3 from the spec
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
				// TODO: add subfield with tag "01"
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

// Step 2. Define the types that matches the ISO 8583 specification
// The `iso8583:"X"` tags tell the library which field number each struct field
// maps to. Field numbers must match exactly what we defined in the spec above.
type AuthorizationRequest struct {
	MTI string `iso8583:"0"`
	// TODO: add PAN and Amount fields
	ProcessingDateTime    string        `iso8583:"4"`
	CurrencyCode          string        `iso8583:"7"`
	CardVerificationValue string        `iso8583:"8"`
	ExpirationDate        string        `iso8583:"9"`
	AcceptorInfo          *AcceptorInfo `iso8583:"10"`
	STAN                  string        `iso8583:"11"`
}

type AcceptorInfo struct {
	// TODO: add MerchantName field
	MCC         string `iso8583:"02"`
	PostalCode  string `iso8583:"03"`
	MerchantURL string `iso8583:"04"`
}

type AuthorizationResponse struct {
	MTI               string `iso8583:"0"`
	ApprovalCode      string `iso8583:"5"`
	AuthorizationCode string `iso8583:"6"`
}

func main() {
	fmt.Println("FTDC Acquirer is running...")

	msg := iso8583.NewMessage(spec)

	// Step 3. Create and populate the AuthorizationRequest message
	authRequest := AuthorizationRequest{
		MTI: "0100",
		// TODO: populate PAN and Amount fields
		ProcessingDateTime:    time.Now().UTC().Format(time.RFC3339),
		CurrencyCode:          "840",
		CardVerificationValue: "1234",
		ExpirationDate:        "2509",
		AcceptorInfo: &AcceptorInfo{
			// TODO: populate MerchantName field
			MCC:         "1234",
			PostalCode:  "12345",
			MerchantURL: "https://fintechdevcon.com",
		},
		STAN: "123456",
	}

	err := msg.Marshal(authRequest)
	if err != nil {
		fmt.Printf("Error marshalling message: %v\n", err)
		return
	}

	packed, err := msg.Pack()

	// Step 4. Inspect the packed message
	// TODO: see the packed message
	// TODO: see fields in human-readable format

	// Step 5. connect to the FTDC Issuer and send the message
	issuerAddr := "localhost:8583"
	conn, err := net.Dial("tcp", issuerAddr)
	if err != nil {
		fmt.Printf("Error connecting to FTDC Issuer: %v\n", err)
		return
	}
	defer conn.Close()

	// Sending message

	// write message length as a 2-bytes header
	// create 2-byte length header (big-endian format)
	// big-endian means most significant byte first
	messageLength := len(packed)
	packedLength := make([]byte, 2)
	packedLength[0] = byte(messageLength >> 8) // high byte
	packedLength[1] = byte(messageLength)      // low byte

	_, err = conn.Write(packedLength)
	if err != nil {
		fmt.Printf("Error sending message length to FTDC Issuer: %v\n", err)
		return
	}

	_, err = conn.Write(packed)
	if err != nil {
		fmt.Printf("Error sending message to FTDC Issuer: %v\n", err)
		return
	}

	fmt.Println("Message sent to FTDC Issuer successfully.")

	// reading response
	responseLength := make([]byte, 2)
	_, err = conn.Read(responseLength)
	if err != nil {
		fmt.Printf("Error reading response length from FTDC Issuer: %v\n", err)
		return
	}

	responseMessageLength := int(responseLength[0])<<8 | int(responseLength[1])
	packed = make([]byte, responseMessageLength)
	_, err = conn.Read(packed)
	if err != nil {
		fmt.Printf("Error reading response from FTDC Issuer: %v\n", err)
		return
	}

	fmt.Printf("Received response from FTDC Issuer:\n%s\n", packed)

	// Step 8. Unpack the response message and print it
	// TODO: see the packed message
	// TODO: unpack the message
	// TODO see fields in human-readable format
}
