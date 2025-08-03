package onemorething

import (
	"github.com/moov-io/iso8583"
	"github.com/moov-io/iso8583/encoding"
	"github.com/moov-io/iso8583/field"
	"github.com/moov-io/iso8583/padding"
	"github.com/moov-io/iso8583/prefix"
	"github.com/moov-io/iso8583/sort"
)

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
		25: field.NewString(&field.Spec{
			Length:      32,
			Description: "Participant Name",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
		}),
	},
}
