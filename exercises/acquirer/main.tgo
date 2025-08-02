package main

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

// 1. Define the ISO 8583 message specification for FTDC Acquirer

// template for ISO 8583 specification
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
		2: field.NewString(&field.Spec{
			Length:      16,
			Description: "Primary Account Number",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		3: field.NewNumeric(&field.Spec{
			Length:      6,
			Description: "Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
		}),
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
				"01": field.NewString(&field.Spec{
					Length:      99,
					Description: "Merchant Name",
					Enc:         encoding.ASCII,
					Pref:        prefix.ASCII.LL,
				}),
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

// 2. Define the AuthorizationRequest struct that matches the ISO 8583 specification

type AuthorizationRequest struct {
	MTI                   string        `iso8583:"0"`
	PAN                   string        `iso8583:"2"`
	Amount                int64         `iso8583:"3"`
	ProcessingDateTime    string        `iso8583:"4"`
	CurrencyCode          string        `iso8583:"7"`
	CardVerificationValue string        `iso8583:"8"`
	ExpirationDate        string        `iso8583:"9"`
	AcceptorInfo          *AcceptorInfo `iso8583:"10"`
	STAN                  string        `iso8583:"11"`
}

type AcceptorInfo struct {
	MerchantName string `iso8583:"01"`
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

func main() {
	fmt.Println("FTDC Acquirer is running...")

	// 4. Create an instance of AuthorizationRequest and marshal it into an ISO 8583 message
	authRequest := AuthorizationRequest{
		MTI:                   "0100",
		PAN:                   "7969689829560386",
		Amount:                199, // $1.99
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

	// 5. Create a new ISO 8583 message using the specification
	// 6. Marshal and pack the message

	msg := iso8583.NewMessage(spec)

	err := msg.Marshal(authRequest)
	if err != nil {
		fmt.Printf("Error marshalling message: %v\n", err)
		return
	}

	packed, err := msg.Pack()
	if err != nil {
		fmt.Printf("Error packing message: %v\n", err)
		return
	}

	// pretty print the packed message
	iso8583.Describe(msg, os.Stdout)

	fmt.Printf("Packed ISO 8583 Message:\n%s\n", packed)

	// 7. connect to the FTDC Issuer and send the message
	conn, err := net.Dial("tcp", "localhost:8583")
	if err != nil {
		fmt.Printf("Error connecting to FTDC Issuer: %v\n", err)
		return
	}
	defer conn.Close()

	// 8. Send the packed message to the FTDC Issuer

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

	_, err = conn.Write(packed)
	if err != nil {
		fmt.Printf("Error sending message to FTDC Issuer: %v\n", err)
		return
	}

	fmt.Println("Message sent to FTDC Issuer successfully.")

	// 9. Read the response from the FTDC Issuer

	responseLength := make([]byte, 2)
	_, err = conn.Read(responseLength)
	if err != nil {
		fmt.Printf("Error reading response length from FTDC Issuer: %v\n", err)
		return
	}

	responseMessageLength := int(responseLength[0])<<8 | int(responseLength[1])
	responsePacked := make([]byte, responseMessageLength)
	_, err = conn.Read(responsePacked)
	if err != nil {
		fmt.Printf("Error reading response from FTDC Issuer: %v\n", err)
		return
	}

	fmt.Printf("Received response from FTDC Issuer:\n%s\n", responsePacked)

	// 10. Unpack the response message and print it

	// let's unpack the message
	unpackedMsg := iso8583.NewMessage(spec)
	err = unpackedMsg.Unpack(responsePacked)
	if err != nil {
		fmt.Printf("Error unpacking message: %v\n", err)
		return
	}

	fmt.Println("Unpacked ISO 8583 Message:")

	iso8583.Describe(unpackedMsg, os.Stdout)
}
