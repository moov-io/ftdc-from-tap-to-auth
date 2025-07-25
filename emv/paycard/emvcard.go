package paycard

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/moov-io/bertlv"
)

// EmvCard represents a payment card.
type EmvCard struct {
	// Contactless read the card using PPSE else PSE
	Contactless             bool
	FileControldInformation []bertlv.TLV
	Applications            []Application
	AIDResponse             AIDResponse
	TwoPayResponse          TwoPayResponse
	GPOResponse             GPOResponse
	SFIResponse             SFIResponse
}

// TwoPayResponse represents the structure of an EMV card 2PAY.SYS.DDF01 response.
type TwoPayResponse struct {
	// DFName is the Dedicated File Name (AID)
	DFName string
	// Applications is a list of applications on the card.
	Applications []Application
	// IssuerPublicKey is the issuer public key.
	IssuerPublicKey []byte
}

// Application represents a payment application on a card.
type Application struct {
	// AID is the application identifier.
	AID []byte
	// Label is the application label.
	Label string
	// Priority is the application priority.
	Priority int
	// Transaction counter ATC
	TransactionCounter int
	// Amount
	Amount float32
}

// AIDResponse represents the parsed response from an EMV card's AID read command.
type AIDResponse struct {
	AID                  string // Application Identifier (AID)
	ApplicationLabel     string // Application Label (Tag 50)
	ApplicationPriority  byte   // Application Priority Indicator (Tag 87)
	LanguagePreference   string // Language Preference (Tag 5F2D)
	IssuerCodeTableIndex byte   // Issuer Code Table Index (Tag 9F11)
	ApplicationVersion   string // Application Version Number (Tag 9F08)
	AIP                  []byte // Application Interchange Profile (AIP - Tag 82)
	AFL                  []byte // Application File Locator (AFL - Tag 94)
	PDOL                 []DOL  // Processing Options Data Object List in Tag Length.
}

// NewEmvCard returns a new EmvCard.
func NewEmvCard(contactless bool) *EmvCard {
	return &EmvCard{
		Contactless: contactless,
	}
}

// ParseEmvCard parses the EMV card data.
func (e *EmvCard) Parse2Pay(data []byte) error {
	// Parse the FCI Template to get the AID
	// After the 2PAY.SYS.DDF01 directory is selected, the card responds with an FCI template. This data structure contains the AIDs of each application registered for contactless payments.
	// The FCI typically uses TLV (Tag-Length-Value) encoding, where each application’s AID is stored as a sequence of bytes under a specific tag, usually 0x4F (the tag for an AID in EMV specifications).

	// todo: handle if the application 2PAY.SYS.DDF01 is not found by checking the status word
	// 6A82 (File not found) or 6A86 (Incorrect P1 or P2 parameter) or 6A88 (Referenced data not found)

	fci, err := bertlv.Decode(data)
	if err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	e.TwoPayResponse = TwoPayResponse{}

	dfname, found := bertlv.FindFirstTag(fci, "84")
	if !found {
		return fmt.Errorf("failed to find DF Name tag 84")
	}
	e.TwoPayResponse.DFName = string(dfname.Value)
	fmt.Printf("DF Name: %s\n", e.TwoPayResponse.DFName)

	e.FileControldInformation = fci

	if err := e.ParseFCIPropertyTemplate(fci); err != nil {
		return fmt.Errorf("failed to parse FCI: %v", err)
	}

	return nil
}

// ParseFCIPropertyTemplate parses the FCI template and populates Application data.
func (e *EmvCard) ParseFCIPropertyTemplate(data []bertlv.TLV) error {
	// Parse the FCI Template to get the AID
	fci, found := bertlv.FindFirstTag(data, "BF0C")
	if !found {
		return fmt.Errorf("failed to find FCI Teamplate tag BF0C")
	}
	for _, applicationTemplate61 := range fci.TLVs {
		application := Application{}
		for _, app := range applicationTemplate61.TLVs {
			// Identify and Extract the AIDs:
			// The NFC reader parses the FCI template to locate each 0x4F tag, which contains the AID for each payment application.
			// Each AID corresponds to a specific card brand, such as Visa, Mastercard, or American Express, and can be mapped using predefined lists of AIDs for each brand.
			switch app.Tag {
			case "4F":
				application.AID = app.Value
			case "50":
				application.Label = string(app.Value)
			case "87":
				application.Priority = int(app.Value[0])
			}
		}
		e.AddApplication(application)

	}
	return nil
}

func (e *EmvCard) AddApplication(app Application) {
	e.Applications = append(e.Applications, app)
}

// NewAIDResponse creates a new AIDResponse instance with parsed data from the EMV response.
func (e *EmvCard) ParseAIDResponse(data []byte) error {
	// parse the AID into a struct
	tlv, err := bertlv.Decode(data)
	if err != nil {
		return fmt.Errorf("failed to decode AID Response: %v", err)
	}
	aidresponse := AIDResponse{}
	// Tag for AID
	tag, found := bertlv.FindFirstTag(tlv, "84")
	if found {
		aidresponse.AID = fmt.Sprintf("%X", tag.Value)
	}

	// Tag for Application Label
	tag, found = bertlv.FindFirstTag(tlv, "50")
	if found {
		aidresponse.ApplicationLabel = string(tag.Value)
	}

	// Tag for Application Priority Indicator
	tag, found = bertlv.FindFirstTag(tlv, "87")
	if found {
		aidresponse.ApplicationPriority = tag.Value[0]
	}

	// Tag for Language Preference
	tag, found = bertlv.FindFirstTag(tlv, "5F2D")
	if found {
		aidresponse.LanguagePreference = fmt.Sprintf("%X", tag.Value)
	}

	// Tag for Issuer Code Table Index
	tag, found = bertlv.FindFirstTag(tlv, "9F11")
	if found {
		aidresponse.IssuerCodeTableIndex = tag.Value[0]
	}

	// Tag for Application Version Number
	tag, found = bertlv.FindFirstTag(tlv, "9F08")
	if found {
		aidresponse.ApplicationVersion = fmt.Sprintf("%X", tag.Value)
	}

	// Tag for AIP
	tag, found = bertlv.FindFirstTag(tlv, "82")
	if found {
		aidresponse.AIP = tag.Value
	}

	// Tag for AFL
	tag, found = bertlv.FindFirstTag(tlv, "94")
	if found {
		aidresponse.AFL = tag.Value
	}

	// Tag for PDOL
	tag, found = bertlv.FindFirstTag(tlv, "9F38")
	if found {
		pdolTagLength, err := ParseDOL(tag.Value)
		if err != nil {
			return fmt.Errorf("error parsing PDOL: %s", err)
		}
		aidresponse.PDOL = pdolTagLength
	}

	e.AIDResponse = aidresponse
	return nil
}

// GPOResponse represents the parsed data from a Get Processing Options response.
type GPOResponse struct {
	AIP                AIP               // Application Interchange Profile (Tag 82)
	Track2Equivalent   *Track2Equivalent // Parsed Track 2 Equivalent Data (Tag 57)
	AFL                []byte            // Application File Locator (Tag 94)
	CardholderName     string            // Cardholder Name (Tag 5F20)
	ApplicationPANSeq  byte              // Application PAN Sequence Number (Tag 5F34)
	IssuerAppData      []byte            // Issuer Application Data (Tag 9F10)
	ARQC               []byte            // Application Cryptogram (Tag 9F26)
	CryptogramInfoData byte              // Cryptogram Information Data (Tag 9F27)
	ATC                []byte            // Application Transaction Counter (Tag 9F36)
	CVMResults         []byte            // Cardholder Verification Method (CVM) Results (Tag 9F6C)
	FormFactor         []byte            // Form Factor Indicator (Tag 9F6E)
	OtherTags          map[string][]byte // Map to store any other tags found
}

// Track2Equivalent represents the parsed data from Track 2 Equivalent Data (Tag 57).
type Track2Equivalent struct {
	PAN           string // Primary Account Number
	Expiration    string // Expiration date in YYMM format
	ServiceCode   string // Service code, usually 3 digits
	Discretionary string // Discretionary data, variable length
}

// parseTrack2Equivalent parses the Track 2 Equivalent Data into a Track2Equivalent struct.
func parseTrack2Equivalent(data []byte) (*Track2Equivalent, error) {
	// Convert data to string for easier parsing
	dataStr := strings.ToUpper(hex.EncodeToString(data))
	track2 := &Track2Equivalent{}

	// Find the separator 'D' indicating end of PAN
	if sepIdx := strings.Index(dataStr, "D"); sepIdx != -1 {
		track2.PAN = dataStr[:sepIdx]

		// Expiration date (4 characters after 'D')
		if len(dataStr) > sepIdx+4 {
			track2.Expiration = dataStr[sepIdx+1 : sepIdx+5]
		}

		// Service code (3 characters after expiration date)
		if len(dataStr) > sepIdx+8 {
			track2.ServiceCode = dataStr[sepIdx+5 : sepIdx+8]
		}

		// Discretionary data (remaining characters)
		if len(dataStr) > sepIdx+8 {
			track2.Discretionary = dataStr[sepIdx+8:]
		}
	} else {
		return nil, fmt.Errorf("separator 'D' not found in Track 2 data")
	}

	return track2, nil
}

func (t *Track2Equivalent) String() string {
	if t == nil {
		return "nil Track2"
	}
	return fmt.Sprintf(
		"\tPAN: %s\n\tExpiration: %s\n\tService Code: %s\n\tDiscretionary Data: %s",
		t.PAN,
		t.Expiration,
		t.ServiceCode,
		t.Discretionary,
	)
}

// String returns a formatted string for displaying GPOResponse contents.
func (g *GPOResponse) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("AIP: %s\n", g.AIP.String()))
	buf.WriteString("Track2 Equivalent: \n")
	if g.Track2Equivalent != nil {
		buf.WriteString(fmt.Sprintf("%s\n", g.Track2Equivalent.String()))
	} else {
		buf.WriteString("<missing>\n")
	}
	buf.WriteString(fmt.Sprintf("Cardholder Name: %s\n", g.CardholderName))
	buf.WriteString(fmt.Sprintf("Application PAN Sequence: %d\n", g.ApplicationPANSeq))
	buf.WriteString(fmt.Sprintf("Issuer Application Data: %s\n", hex.EncodeToString(g.IssuerAppData)))
	buf.WriteString(fmt.Sprintf("ARQC: %s\n", hex.EncodeToString(g.ARQC)))
	buf.WriteString(fmt.Sprintf("Cryptogram Information Data: %d\n", g.CryptogramInfoData))
	buf.WriteString(fmt.Sprintf("ATC: %s\n", hex.EncodeToString(g.ATC)))
	buf.WriteString(fmt.Sprintf("CVM Results: %s\n", hex.EncodeToString(g.CVMResults)))
	buf.WriteString(fmt.Sprintf("Form Factor Indicator: %s\n", hex.EncodeToString(g.FormFactor)))
	buf.WriteString(fmt.Sprintf("AFL: %s\n", hex.EncodeToString(g.AFL)))
	buf.WriteString(fmt.Sprintf("Other Tags: %v\n", g.OtherTags))

	return buf.String()
}

// ParseGPOResponse takes the GPO response bytes and parses them into a GPOResponse struct.
func (e *EmvCard) ParseGPOResponse(data []byte) error {

	gporesponse := GPOResponse{
		OtherTags: make(map[string][]byte),
	}

	// Iterate through the response data to parse tags

	tlv, err := bertlv.Decode(data)
	if err != nil {
		return fmt.Errorf("failed to decode AID Response: %v", err)
	}

	response, found := bertlv.FindFirstTag(tlv, "77")
	if !found {
		response, found = bertlv.FindFirstTag(tlv, "80")
		if !found {
			return fmt.Errorf("failed to find DF Name tag 77 or 80")
		}
		// 80 Tag (Constructed TLV): Indicates the response contains a single TLV
		// 80 [length] [data]
		emvData, err := parseEMVTag80(data)
		if err != nil {
			return fmt.Errorf("Failed to parse EMV response: %v", err)
		}

		// Display parsed data
		fmt.Printf("Parsed EMV Data:\n")
		fmt.Printf("AIP: %s\n", emvData.AIP)
		fmt.Println("AFL:")
		for _, afl := range emvData.AFL {
			fmt.Printf("  FileID: 0x%X, StartRecord: %d, EndRecord: %d, NumberOfSFI: 0x%X\n",
				afl.FileID, afl.StartRecord, afl.EndRecord, afl.NumberOfSFI)
		}
		return nil

	}

	// 77 Tag (Constructed TLV): Indicates the response contains multiple TLVs
	for _, tag := range response.TLVs {
		switch tag.Tag {
		case "82":
			err := e.ParseAIP(tag.Value)
			if err != nil {
				return fmt.Errorf("failed to parse AIP: %v", err)
			}
		case "57":
			track2, err := parseTrack2Equivalent(tag.Value)
			if err != nil {
				return fmt.Errorf("failed to parse Track 2 Equivalent Data: %v", err)
			}
			gporesponse.Track2Equivalent = track2
		case "94":
			gporesponse.AFL = tag.Value
		case "5F20":
			gporesponse.CardholderName = string(tag.Value)
		case "5F34":
			if len(tag.Value) > 0 {
				gporesponse.ApplicationPANSeq = tag.Value[0]
			}
		case "9F10":
			gporesponse.IssuerAppData = tag.Value
		case "9F26":
			gporesponse.ARQC = tag.Value
		case "9F27":
			if len(tag.Value) > 0 {
				gporesponse.CryptogramInfoData = tag.Value[0]
			}
		case "9F36":
			gporesponse.ATC = tag.Value
		case "9F6C":
			gporesponse.CVMResults = tag.Value
		case "9F6E":
			gporesponse.FormFactor = tag.Value
		default:
			// Store any other tags in OtherTags map
			gporesponse.OtherTags[tag.Tag] = tag.Value
		}
	}
	e.GPOResponse = gporesponse
	return nil
}

// EMVTag80 represents the parsed AIP and AFL from the 80 tag response.
type EMVTag80 struct {
	AIP string   // Application Interchange Profile (2 bytes, hex)
	AFL []AFLSet // Application File Locator (variable length, parsed into sets)
}

// AFLSet represents a single entry in the AFL.
type AFLSet struct {
	FileID      byte // File Identifier (1 byte)
	StartRecord byte // Starting record number (1 byte)
	EndRecord   byte // Ending record number (1 byte)
	NumberOfSFI byte // Number of Short File Identifiers (SFI)
}

// parseEMVTag80 parses a GPO response with the 80 tag into an EMVData struct.
func parseEMVTag80(data []byte) (*EMVTag80, error) {
	if len(data) < 4 || data[0] != 0x80 {
		return nil, fmt.Errorf("invalid EMV response, missing 80 tag")
	}

	// Check for status word (90 00) at the end
	if len(data) < 2 || data[len(data)-2] != 0x90 || data[len(data)-1] != 0x00 {
		return nil, fmt.Errorf("invalid EMV response, missing status word 90 00")
	}

	// Extract the length byte and verify size (excluding status word)
	length := data[1]
	expectedLength := int(2 + length)
	if len(data) != expectedLength+2 { // +2 for status word
		return nil, fmt.Errorf("length mismatch: expected %d bytes (plus 2 for status word), got %d", expectedLength, len(data)-2)
	}

	// Extract AIP (2 bytes) and AFL (remaining bytes, excluding status word)
	payload := data[2 : len(data)-2]
	if len(payload) < 2 {
		return nil, fmt.Errorf("invalid payload length for AIP and AFL")
	}

	aip := payload[:2]
	afl := payload[2:]

	// Parse AFL into AFLSet structs
	aflSets, err := parseAFL(afl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AFL: %v", err)
	}

	// Return the parsed EMVData
	return &EMVTag80{
		AIP: hex.EncodeToString(aip),
		AFL: aflSets,
	}, nil
}

// parseAFL parses the AFL portion of the response into a slice of AFLSet.
func parseAFL(data []byte) ([]AFLSet, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("AFL length is not a multiple of 4")
	}

	var aflSets []AFLSet
	reader := bytes.NewReader(data)

	for reader.Len() > 0 {
		var set AFLSet
		err := binary.Read(reader, binary.BigEndian, &set)
		if err != nil {
			return nil, fmt.Errorf("failed to read AFL set: %v", err)
		}
		aflSets = append(aflSets, set)
	}

	return aflSets, nil
}

// AIP represents the Application Interchange Profile in an EMV card,
// with each field indicating a specific card capability.
type AIP struct {
	OfflineDataAuthentication  bool // Bit 8 of byte 1
	CardholderVerification     bool // Bit 7 of byte 1
	TerminalRiskManagement     bool // Bit 6 of byte 1
	IssuerAuthentication       bool // Bit 5 of byte 1
	CombinedDataAuthentication bool // Bit 4 of byte 1
}

func (a AIP) String() string {
	return fmt.Sprintf(
		"Offline Data Authentication: %t\nCardholder Verification: %t\nTerminal Risk Management: %t\nIssuer Authentication: %t\nCombined Data Authentication: %t",
		a.OfflineDataAuthentication,
		a.CardholderVerification,
		a.TerminalRiskManagement,
		a.IssuerAuthentication,
		a.CombinedDataAuthentication,
	)
}

// ParseAIP takes the AIP bytes from the EMV response and returns an AIP struct
// with each field set according to the respective bit in the AIP.
func (e *EmvCard) ParseAIP(aipBytes []byte) error {
	if len(aipBytes) < 2 {
		return fmt.Errorf("invalid AIP length: expected at least 2 bytes, got %d", len(aipBytes))
	}

	aip := AIP{
		OfflineDataAuthentication:  aipBytes[0]&0x80 != 0, // Bit 8 of byte 1
		CardholderVerification:     aipBytes[0]&0x40 != 0, // Bit 7 of byte 1
		TerminalRiskManagement:     aipBytes[0]&0x20 != 0, // Bit 6 of byte 1
		IssuerAuthentication:       aipBytes[0]&0x10 != 0, // Bit 5 of byte 1
		CombinedDataAuthentication: aipBytes[0]&0x08 != 0, // Bit 4 of byte 1
	}
	e.GPOResponse.AIP = aip
	return nil
}

// SFIResponse represents the parsed SFI (Short File Identifier) response.
type SFIResponse struct {
	SFI          byte // Short File Identifier
	RecordNumber byte // Record Number
	// CTQ - 9F6C Card Transaction Qualifier (CTQ) is a set of data elements, set by the card issuer, that determines the actions taken at the point of sale (POS) during a transaction, such as whether a cardholder verification method is required or if the transaction can be processed offline
	CTQ []byte // Card Transaction Qualifier
	// PCVC3 - 9F62 PCVC3 (Precomputed Card Verification Code 3) for Track 1 is used in the dynamic card authentication process
	PCVC3     []byte // Precomputed Card Verification Code 3
	OtherTags map[string][]byte
}

// ParseSFI parses the SFI (Short File Identifier) from the EMV response.
func (e *EmvCard) ParseSFI(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("invalid SFI length: expected at least 2 bytes, got %d", len(data))
	}

	sfiresponse := SFIResponse{
		OtherTags: make(map[string][]byte),
	}
	// Extract the SFI (Short File Identifier) from the first byte
	sfi := data[0]
	// Extract the Record Number from the second byte
	recordNumber := data[1]
	// Store the SFI and Record Number in the SFIResponse struct
	sfiresponse.SFI = sfi
	sfiresponse.RecordNumber = recordNumber

	tlv, err := bertlv.Decode(data)
	if err != nil {
		return fmt.Errorf("failed to decode SFI response: %v", err)
	}
	// Iterate through the TLV data to extract other tags
	response, found := bertlv.FindFirstTag(tlv, "70")
	if !found {
		return fmt.Errorf("failed to find read record response message template tag 70")
	}
	for _, tag := range response.TLVs {
		switch tag.Tag {
		case "9F6C":
			sfiresponse.CTQ = tag.Value
		case "9F62":
			sfiresponse.PCVC3 = tag.Value
		default:
			sfiresponse.OtherTags[tag.Tag] = tag.Value
		}
	}
	e.SFIResponse = sfiresponse

	fmt.Printf("=> 💳 Reading SFI %d, Record %d : %X\n", sfi, recordNumber, data)

	return nil
}
