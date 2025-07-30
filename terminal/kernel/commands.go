package kernel

// PPSE (Payment System Environment) - the "directory" of payment applications on the card
// It's like asking the card "what payment apps do you have?"
// "2PAY.SYS.DDF01"
var PPSE_AID = []byte("2PAY.SYS.DDF01")

// SELECT command - used to select applications or files on the card
// This is like "opening" an application or asking to see a specific file
func NewSelectCommand(aid []byte) APDUCommand {
	return APDUCommand{
		CLA:  0x00, // Standard class byte
		INS:  0xA4, // SELECT instruction
		P1:   0x04, // Select by DF (Dedicated File) name
		P2:   0x00, // First or only occurrence
		Data: aid,  // The AID we want to select
		Le:   ptrByte(0),
	}
}

// READ RECORD command - reads a specific record from a file on the card
// This is used to extract cardholder data like PAN, name, expiration date
func NewReadRecordCommand(recordNumber byte, sfi byte) APDUCommand {
	return APDUCommand{
		CLA: 0x00,              // Standard class
		INS: 0xB2,              // READ RECORD instruction
		P1:  recordNumber,      // Record number (1, 2, 3, etc.)
		P2:  (sfi << 3) | 0x04, // SFI in upper 5 bits + reference control (mode 4)
		Le:  ptrByte(0),        // Return all available data
	}
}
