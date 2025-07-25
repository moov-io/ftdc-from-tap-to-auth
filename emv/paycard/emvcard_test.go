package paycard

import (
	"encoding/hex"
	"testing"
)

// Parse2Pay parses the EMV card data.
func TestParse2Pay(t *testing.T) {
	// payResponse := []byte({0x6f, 0x30, ...})
	payResponse, err := hex.DecodeString("6F30840E325041592E5359532E4444463031A51EBF0C1B61194F07A0000000031010500B56495341204352454449548701019000")
	if err != nil {
		t.Errorf("Error decoding 2PAY response: %v", err)
	}
	emvCard := NewEmvCard(true)

	err = emvCard.Parse2Pay(payResponse)
	if err != nil {
		t.Errorf("Error parsing 2PAY response: %v", err)
	}

	if emvCard.TwoPayResponse.DFName != PPSE {
		t.Errorf("DFName not parsed correctly wanted: %s got: %s", PPSE, emvCard.AIDResponse.AID)
	}

}

// ParseAIDResponse parses the AID response.
func TestParseAIDresponse(t *testing.T) {

	AIDResponse, err := hex.DecodeString("6F468407A0000000031010A53B500B56495341204352454449548701015F2D02656E9F38189F66049F02069F03069F1A0295055F2A029A039C019F3704BF0C089F5A051108400840")
	if err != nil {
		t.Errorf("Error decoding AID response: %v", err)
	}

	emvCard := NewEmvCard(true)

	err = emvCard.ParseAIDResponse(AIDResponse)
	if err != nil {
		t.Errorf("Error parsing AID response: %v", err)
	}
	dfExpected := "A0000000031010"
	if emvCard.AIDResponse.AID != dfExpected {
		t.Errorf("DFName not parsed correctly wanted: %s got: %s", dfExpected, emvCard.AIDResponse.AID)
	}
}

func TestParseGPOResponse(t *testing.T) {
	gpodata, err := hex.DecodeString("77598202200057134147202500716749D26072011010041301051F5F200F43415244484F4C4445522F564953415F3401019F100706021203A000009F2608D0C669EEB70C58DD9F2701809F360200699F6C0200009F6E04207000009000")
	if err != nil {
		t.Errorf("Error decoding GPO response: %v", err)
	}

	emvCard := NewEmvCard(true)

	err = emvCard.ParseGPOResponse(gpodata)
	if err != nil {
		t.Errorf("Error parsing GPO response: %v", err)
	}

	want := "CARDHOLDER/VISA"
	if emvCard.GPOResponse.CardholderName != want {
		t.Errorf("GPO Response %s\n", emvCard.GPOResponse.String())
		//t.Errorf("CardholderName not parsed correctly wanted: %s got: %X", want, emvCard.GPOResponse.CardholderName)
	}
}

func TestParseAIP(t *testing.T) {
	aipdata, err := hex.DecodeString("8000")
	if err != nil {
		t.Errorf("Error decoding AIP response: %v", err)
	}

	emvCard := NewEmvCard(true)

	err = emvCard.ParseAIP(aipdata)
	if err != nil {
		t.Errorf("Error parsing AIP response: %v", err)
	}

	want := true
	if emvCard.GPOResponse.AIP.OfflineDataAuthentication != want {
		t.Errorf("AIP not parsed correctly wanted: %t got: %t", want, emvCard.GPOResponse.AIP.OfflineDataAuthentication)
	}
}

func TestParseGPO80(t *testing.T) {
	gpo80data, err := hex.DecodeString("800A3C001001010114020303")
	emvcard := NewEmvCard(true)

	err = emvcard.ParseGPOResponse(gpo80data)
	if err != nil {
		t.Errorf("Error decoding GPO 80 response: %v", err)
	}

	//	want := "3C001001010114020303"
	//	if emvcard.GPOResponse. != want {
	//		t.Errorf("GPO80 not parsed correctly wanted: %s got: %s", want, emvcard.GPOResponse.GPO80)
	//	}

}
