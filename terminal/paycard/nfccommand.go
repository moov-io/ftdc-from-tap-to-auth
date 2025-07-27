package paycard

// PPSE is a byte slice representing the string "2PAY.SYS.DDF01"
var PPSE = "2PAY.SYS.DDF01"

// PSE is a byte slice representing the string "1PAY.SYS.DDF01"
var PSE = "1PAY.SYS.DDF01"

// Command represents an EMV APDU command with fields for the class, instruction, and parameters.
type Command struct {
	// Cla Instruction Class – indicates the type of command, e.g. interindustry or proprietary
	Cla int
	// Ins Instruction code – indicates the specific command, e.g. write data
	//
	Ins int
	// P1 Parameter 1 – further specifies the command, e.g. offset into file to write the data
	P1 int
	// P2 Parameter 2 – further specifies the command, e.g. number of bytes to write
	P2 int
	// Lc Length of the command data
	Lc int
	// Data Command data
	Data []byte
	// Le Length of the expected response data
	// todo how to handle this as it is normally zero
	Le int
}

// Define each command as a constant Command.
var (
	Select      = Command{Cla: 0x00, Ins: 0xA4, P1: 0x04, P2: 0x00}
	ReadRecord  = Command{Cla: 0x00, Ins: 0xB2, P1: 0x00, P2: 0x00}
	GPO         = Command{Cla: 0x80, Ins: 0xA8, P1: 0x00, P2: 0x00}
	GetData     = Command{Cla: 0x80, Ins: 0xCA, P1: 0x00, P2: 0x00}
	GetResponse = Command{Cla: 0x00, Ins: 0x0C, P1: 0x00, P2: 0x00}
)

var (
	// SelectPPSE is the SELECT command for the PPSE directory.
	SelectPPSE = Command{Cla: 0x00, Ins: 0xA4, P1: 0x04, P2: 0x00, Lc: len(PPSE), Data: []byte(PPSE)}
	// SelectPSE is the SELECT command for the PSE directory.
	SelectPSE = Command{Cla: 0x00, Ins: 0xA4, P1: 0x04, P2: 0x00, Lc: len(PSE), Data: []byte(PSE)}
)

// Bytes returns the byte representation of the command.
func (c Command) Bytes() []byte {
	bytes := []byte{byte(c.Cla), byte(c.Ins), byte(c.P1), byte(c.P2)}
	if c.Lc != 0 {
		bytes = append(bytes, byte(c.Lc))
	}
	if c.Data != nil {
		bytes = append(bytes, c.Data...)
	}
	bytes = append(bytes, byte(c.Le))
	return bytes
}

// SelectAID returns a SELECT Command for the given AID.
func SelectAID(aid []byte) Command {
	return Command{Cla: 0x00, Ins: 0xA4, P1: 0x04, P2: 0x00, Lc: len(aid), Data: aid}
}

// GetProcessingOptions constructs the GET PROCESSING OPTIONS (GPO) command according to EMV specifications.
// This command initiates the transaction processing on the card and retrieves the Application Interchange Profile (AIP)
// and Application File Locator (AFL).
//
// Example usage:
// pdolData := terminal.BuildPDOLData(&session, emvCard.AIDResponse.PDOL)
// command := GetProcessingOptions(pdolData)
//
// Expected Response Format:
// Success response will be in one of two formats:
//
//  1. Format 1 (Tag 80):
//     80 XX <AIP 2 bytes> <AFL remaining bytes>
//
//  2. Format 2 (Tag 77):
//     77 XX <Template with AIP and AFL as separate objects>
//     - 82 02 <AIP 2 bytes>
//     - 94 XX <AFL multiple bytes>
//
// Common response status words:
// 9000: Success
// 6700: Wrong length
// 6985: Conditions of use not satisfied
// 6984: Random number is invalid
// 6A86: Incorrect P1/P2 parameters
func GetProcessingOptions(pdol []byte) []byte {
	// If no PDOL data is provided, return a simple GPO command
	if len(pdol) == 0 {
		return []byte{
			0x80, // CLA: '80' indicates proprietary class for payment systems
			0xA8, // INS: 'A8' is the instruction code for GET PROCESSING OPTIONS
			0x00, // P1: Parameter 1, always '00' for GPO
			0x00, // P2: Parameter 2, always '00' for GPO
			0x02, // Le: Expected length of response, '02' for header length
			0x83, // Command template tag
			0x00, // Length of PDOL data
			0x00, // Expected length of response data
		}
		// 80 A8 00 00 02 83 00 00

	}

	// Start building the command with mandatory header
	command := []byte{
		0x80, // CLA: '80' for proprietary class (payment systems)
		0xA8, // INS: 'A8' for GET PROCESSING OPTIONS
		0x00, // P1: '00' - no specific parameters needed
		0x00, // P2: '00' - no specific parameters needed
	}

	// Construct PDOL related data object
	// Format: <Tag><Length><Value>
	pdolWrapper := []byte{
		0x83,            // Tag '83' - Command template for GPO (specified in EMV Book 3)
		byte(len(pdol)), // Length of the PDOL data
	}
	pdolWrapper = append(pdolWrapper, pdol...) // Add the actual PDOL data

	// Calculate and append the total length of data field (Lc)
	// This includes the length of: PDOL wrapper tag (1) + PDOL length byte (1) + PDOL data
	command = append(command, byte(len(pdolWrapper)))

	// Append the wrapped PDOL data
	command = append(command, pdolWrapper...)

	// Append Le byte
	// '00' indicates that we want to receive all available response data
	command = append(command, 0x00)

	// The final command structure is:
	// +-----+-----+-----+-----+-----+---------------+-----+
	// | CLA | INS | P1  | P2  | Lc  | PDOL Data    | Le  |
	// | 80  | A8  | 00  | 00  | XX  | 83 LL <data> | 00  |
	// +-----+-----+-----+-----+-----+---------------+-----+
	//
	// Where:
	// - CLA (1 byte): Class of instruction (80 for proprietary)
	// - INS (1 byte): Instruction code (A8 for GPO)
	// - P1 (1 byte): Parameter 1 (00)
	// - P2 (1 byte): Parameter 2 (00)
	// - Lc (1 byte): Length of data field
	// - PDOL Data:
	//   - 83: Command template tag
	//   - LL: Length of PDOL data
	//   - data: Actual PDOL values
	// - Le (1 byte): Expected length of response (00)

	return command
}

// CalculateSFI shifts the SFI left by 3 bits to get the actual SFI record value.
func CalculateSFI(sfi byte) byte {
	return (sfi << 3) | 0x04 // OR with 0x04 as per EMV standard for Read Record
}

// GenerateReadCommands generates READ RECORD commands based on AFL input.
func GenerateReadCommands(afl []byte) [][]byte {
	var commands [][]byte
	for i := 0; i < len(afl); i += 4 {
		sfi := afl[i] >> 3 // Extract SFI
		firstRecord := afl[i+1]
		lastRecord := afl[i+2]

		for record := firstRecord; record <= lastRecord; record++ {
			cmd := []byte{0x00, 0xB2, record, CalculateSFI(sfi), 0x00}
			commands = append(commands, cmd)
		}
	}
	return commands
}

/**

Here is a list of common response codes that can be returned from a GET PROCESSING OPTIONS (GPO) command in an EMV transaction, along with their definitions:

Common Response Codes for GET PROCESSING OPTIONS (GPO)
Successful Response:
9000: Success – The command has been successfully executed, and the response data can be used by the terminal for further processing.
Error Responses:
6283: State of Non-Volatile Memory Changed – The command was executed but some part of non-volatile memory has changed, which may require attention or updates.
6300: Authentication of Cryptogram Failed – Indicates a failure in cryptographic checks related to the application cryptogram.
6581: Memory Failure – General memory failure that might be related to storage or memory issues on the card.
6700: Wrong Length – Indicates that the length of the data field in the command is incorrect or not supported.
6981: Command Incompatible with File Structure – The command sent does not align with the card’s internal file structure.
6982: Security Status Not Satisfied – Security conditions are not met, which could indicate that authentication or other necessary steps are incomplete.
6983: Authentication Method Blocked – The card’s authentication mechanism has been blocked, possibly due to repeated failed authentication attempts.
6985: Conditions of Use Not Satisfied – The conditions for using the card or executing the command are not met. This may occur if the card has usage restrictions, or the command sequence was incorrect.
6986: Command Not Allowed (No Current EF) – The command cannot be performed because there is no current elementary file (EF) selected.
6A80: Incorrect Parameters in the Data Field – The command includes invalid or improperly formatted data.
6A81: Function Not Supported – The card does not support the requested function.
6A82: File or Application Not Found – The requested file or application is not present on the card.
6A83: Record Not Found – The record referred to in the command cannot be found.
6A84: Not Enough Memory Space – Insufficient space in memory for the requested operation.
6A85: Lc Inconsistent with TLV Structure – The length of the command data field does not match the TLV structure.
6A86: Incorrect Parameters (P1-P2) – The parameters P1 and P2 in the command are incorrect.
6A88: Referenced Data Not Found – The referenced data in the command cannot be found.
6D00: Invalid Instruction Code – The instruction code (INS) in the command is invalid or not supported by the card.
6E00: Invalid Class Code – The class byte (CLA) in the command is invalid or not recognized by the card.
6F00: Unknown Error – A generic error indicating an unspecified problem with processing the command.
**/
