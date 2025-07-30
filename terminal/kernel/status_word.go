package kernel

import "fmt"

// StatusWordInfo contains human-readable information about APDU status words
type StatusWordInfo struct {
	Code        uint16
	Name        string
	Description string
	Category    string // Success, Warning, Error, Security
}

// Common EMV status words with human-readable descriptions
var statusWords = map[uint16]StatusWordInfo{
	// Success responses
	0x9000: {0x9000, "Success", "Command completed successfully", "Success"},

	// Warning responses (61xx, 62xx, 63xx)
	0x6100: {0x6100, "More Data Available", "More data available, use GET RESPONSE", "Warning"},
	0x6200: {0x6200, "Warning - No Info", "Warning condition, no information given", "Warning"},
	0x6281: {0x6281, "Part of Data Corrupted", "Part of returned data may be corrupted", "Warning"},
	0x6282: {0x6282, "EOF Before Reading", "End of file reached before reading expected number of bytes", "Warning"},
	0x6283: {0x6283, "Selected File Deactivated", "Selected file is deactivated", "Warning"},
	0x6284: {0x6284, "File Control Info Error", "File control information not formatted according to standard", "Warning"},
	0x6300: {0x6300, "Authentication Failed", "Authentication failed", "Security"},
	0x63C0: {0x63C0, "PIN Verification Failed", "PIN verification failed, no tries left", "Security"},
	0x63C1: {0x63C1, "PIN Verification Failed", "PIN verification failed, 1 try left", "Security"},
	0x63C2: {0x63C2, "PIN Verification Failed", "PIN verification failed, 2 tries left", "Security"},
	0x63C3: {0x63C3, "PIN Verification Failed", "PIN verification failed, 3 tries left", "Security"},

	// Execution errors (64xx, 65xx)
	0x6400: {0x6400, "Execution Error", "Execution error, no information given", "Error"},
	0x6401: {0x6401, "Immediate Response Required", "Execution error, immediate response required", "Error"},
	0x6500: {0x6500, "Memory Failure", "Execution error, memory failure", "Error"},

	// Checking errors (66xx through 6Fxx)
	0x6600: {0x6600, "Security Issue", "Execution error, security issue", "Security"},
	0x6700: {0x6700, "Wrong Length", "Wrong length in Lc field", "Error"},
	0x6800: {0x6800, "Unsupported Function", "Function not supported in CLA", "Error"},
	0x6881: {0x6881, "Logical Channel Not Supported", "Logical channel not supported", "Error"},
	0x6882: {0x6882, "Security Messaging Not Supported", "Secure messaging not supported", "Security"},
	0x6900: {0x6900, "Command Not Allowed", "Command not allowed, no information given", "Error"},
	0x6981: {0x6981, "Command Incompatible", "Command incompatible with file structure", "Error"},
	0x6982: {0x6982, "Security Status Not Satisfied", "Security status not satisfied", "Security"},
	0x6983: {0x6983, "Authentication Method Blocked", "Authentication method blocked", "Security"},
	0x6984: {0x6984, "Reference Data Not Usable", "Reference data not usable", "Security"},
	0x6985: {0x6985, "Conditions Not Satisfied", "Conditions of use not satisfied", "Error"},
	0x6986: {0x6986, "Command Not Allowed", "Command not allowed (no current EF)", "Error"},
	0x6987: {0x6987, "Expected Secure Messaging", "Expected secure messaging data objects missing", "Security"},
	0x6988: {0x6988, "Incorrect Secure Messaging", "Incorrect secure messaging data objects", "Security"},
	0x6A00: {0x6A00, "Wrong Parameters", "Wrong parameters P1-P2", "Error"},
	0x6A80: {0x6A80, "Incorrect Data", "Incorrect parameters in command data field", "Error"},
	0x6A81: {0x6A81, "Function Not Supported", "Function not supported", "Error"},
	0x6A82: {0x6A82, "File Not Found", "File or application not found", "Error"},
	0x6A83: {0x6A83, "Record Not Found", "Record not found", "Error"},
	0x6A84: {0x6A84, "Not Enough Memory", "Not enough memory space in the file", "Error"},
	0x6A85: {0x6A85, "Incorrect TLV", "Nc inconsistent with TLV structure", "Error"},
	0x6A86: {0x6A86, "Incorrect P1-P2", "Incorrect parameters P1-P2", "Error"},
	0x6A87: {0x6A87, "Nc Inconsistent with P1-P2", "Nc inconsistent with parameters P1-P2", "Error"},
	0x6A88: {0x6A88, "Referenced Data Not Found", "Referenced data or reference data not found", "Error"},
	0x6A89: {0x6A89, "File Already Exists", "File already exists", "Error"},
	0x6A8A: {0x6A8A, "DF Name Already Exists", "DF name already exists", "Error"},
	0x6B00: {0x6B00, "Wrong Parameters", "Wrong parameters P1-P2", "Error"},
	0x6C00: {0x6C00, "Wrong Le Field", "Wrong Le field", "Error"},
	0x6CFF: {0x6CFF, "Wrong Le Field", "Wrong Le field, exact length is 255", "Error"},
	0x6D00: {0x6D00, "Instruction Not Supported", "Instruction code not supported or invalid", "Error"},
	0x6E00: {0x6E00, "Class Not Supported", "Class not supported", "Error"},
	0x6F00: {0x6F00, "No Precise Diagnosis", "No precise diagnosis", "Error"},

	// EMV-specific status words
	0x9100: {0x9100, "Terminal Risk Management", "Terminal risk management was performed", "Warning"},
	0x9101: {0x9101, "Issuer Authentication Failed", "Issuer authentication failed", "Security"},
	0x9102: {0x9102, "Script Processing Failed", "Script processing failed", "Error"},
	0x9110: {0x9110, "PIN Try Limit Exceeded", "PIN Try Limit exceeded", "Security"},
	0x9202: {0x9202, "Service Not Allowed", "Service not allowed for card product", "Error"},
	0x9210: {0x9210, "PIN Required", "PIN required", "Security"},
	0x9220: {0x9220, "PIN Block Format Error", "PIN block format error", "Security"},
	0x9900: {0x9900, "Terminal Application Error", "Terminal application error", "Error"},
}

// GetStatusWordInfo returns human-readable information about a status word
func GetStatusWordInfo(statusWord uint16) StatusWordInfo {
	if info, exists := statusWords[statusWord]; exists {
		return info
	}

	// Handle variable status words
	switch {
	case statusWord >= 0x61F0 && statusWord <= 0x61FF:
		return StatusWordInfo{
			Code:        statusWord,
			Name:        "More Data Available",
			Description: fmt.Sprintf("More data available, %d bytes can be read with GET RESPONSE", statusWord&0xFF),
			Category:    "Warning",
		}
	case statusWord >= 0x6C00 && statusWord <= 0x6CFF:
		return StatusWordInfo{
			Code:        statusWord,
			Name:        "Wrong Le Field",
			Description: fmt.Sprintf("Wrong Le field, exact length expected is %d", statusWord&0xFF),
			Category:    "Error",
		}
	case statusWord >= 0x63C0 && statusWord <= 0x63CF:
		triesLeft := statusWord & 0x0F
		return StatusWordInfo{
			Code:        statusWord,
			Name:        "PIN Verification Failed",
			Description: fmt.Sprintf("PIN verification failed, %d tries left", triesLeft),
			Category:    "Security",
		}
	}

	// Unknown status word
	return StatusWordInfo{
		Code:        statusWord,
		Name:        "Unknown Status",
		Description: fmt.Sprintf("Unknown status word: %04X", statusWord),
		Category:    "Error",
	}
}
