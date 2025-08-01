# Templates for building acquirer tasks

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
			Enc:         encoding.Binary,
			Pref:        prefix.Binary.Fixed,
		}),
	},
}
```

composite field example:

```
field.NewComposite(&field.Spec{
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
    },
}),
```


