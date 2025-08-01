# Fintech DevCon ISO 8583 Specification

This specification defines a simplified ISO 8583 message format for educational and demonstration purposes at [Fintech DevCon](https://fintechdevcon.io).

## Encoding Notes

- **ASCII**: Standard ASCII character encoding for all alphanumeric and special characters
- **Binary**: Raw binary data encoding
- **Fixed Length**: Field has a predetermined, unchanging length
- **Variable Length**: Field length is indicated by a length prefix
  - **LL**: 2-digit decimal length indicator
  - **LLL**: 3-digit decimal length indicator

## Message Types

### 0100 / 0110 - Authorization Request / Response

| Field | Element Name | Req/Resp | Format | Length | Description |
|-------|--------------|---------|---------|---------|-------------|
| 0 | Message Type Indicator | Req | ANS | 4 | Fixed: "0100" |
| 1 | Bitmap | Req / Res | B, HEX | 8 | Presence indicator |
| 2 | Primary Account Number (PAN) | Req | ANS | 16 | Card number |
| 3 | Amount | Req | N | 6 | Transaction amount |
| 4 | Transmission Date & Time | Req | ANS | 20 | Message timestamp |
| 5 | Approval Code | Resp | ANS | 2 | Authorization result |
| 6 | Authorization Code | Resp | ANS | 6 | Issuer auth code |
| 7 | Currency | Req | ANS | 3 | Currency code |
| 8 | Card Verification Value (CVV) | Req | ANS | 4 | Security code |
| 9 | Card Expiration Date | Req | ANS | 4 | Card expiry |
| 10 | Acceptor Information | Req | COMP | VAR | Merchant details |
| 11 | Systems Trace Audit Number (STAN) | Req / Res | ANS | 6 | Trace number |
| 55 | Chip Data | Req | B | 999 | EMV chip data |

### 0110 - Authorization Response

**Legend:**
- Req = Request, Resp = Response
- ANS = Alphanumeric and Special, N = Numeric, B = Binary, COMP = Composite
- VAR = Variable length

## Data Elements

### Field 0 - Message Type Indicator
- **Type**: String
- **Length**: 4 characters (fixed)
- **Encoding**: ASCII
- **Description**: Identifies the message type and function

### Field 1 - Bitmap
- **Type**: Bitmap
- **Length**: 8 bytes (fixed)
- **Encoding**: HEX(ASCII)
- **Description**: Indicates which data elements are present in the message

### Field 2 - Primary Account Number (PAN)
- **Type**: String
- **Length**: 16 characters (fixed)
- **Encoding**: ASCII
- **Description**: The primary account number associated with the card

### Field 3 - Amount
- **Type**: Numeric
- **Length**: 6 digits (fixed)
- **Encoding**: ASCII
- **Padding**: Left-padded with zeros
- **Description**: Transaction amount

### Field 4 - Transmission Date & Time
- **Type**: String
- **Length**: 20 characters (fixed)
- **Encoding**: ASCII
- **Description**: Date and time when the message was transmitted
- **Format**: YYYY-MM-DDTHH:MM:SSZ (UTC time), in Go - time.RFC3339. Example: `2009-11-10T23:00:00Z`

### Field 5 - Approval Code
- **Type**: String
- **Length**: 2 characters (fixed)
- **Encoding**: ASCII
- **Description**: Authorization approval code

**List of approval codes:**
* 00 - Approved
* 05 - Declined
* 10 - Invalid Request
* 14 - Invalid/Unknown Card
* 51 - Insufficient Funds
* 99 - System Error


### Field 6 - Authorization Code
- **Type**: String
- **Length**: 6 characters (fixed)
- **Encoding**: ASCII
- **Description**: Authorization code from the issuer

### Field 7 - Currency
- **Type**: String
- **Length**: 3 characters (fixed)
- **Encoding**: ASCII
- **Description**: Transaction currency code

### Field 8 - Card Verification Value (CVV)
- **Type**: String
- **Length**: 4 characters (fixed)
- **Encoding**: ASCII
- **Description**: Card verification value for security

### Field 9 - Card Expiration Date
- **Type**: String
- **Length**: 4 characters (fixed)
- **Encoding**: ASCII
- **Description**: Card expiration date, formatted as YYMM (e.g., 2401 for January 2024)

### Field 10 - Acceptor Information
- **Type**: Composite
- **Length**: Variable (up to 999 characters)
- **Encoding**: ASCII
- **Length Prefix**: LLL (3-digit length indicator)
- **Tag Format**: 2-digit ASCII tags, sorted by integer value
- **Description**: Information about the merchant/acceptor

#### Subfields:

##### Tag 01 - Merchant Name
- **Type**: String
- **Length**: Variable (up to 99 characters)
- **Encoding**: ASCII
- **Length Prefix**: LL (2-digit length indicator)
- **Description**: Name of the merchant

##### Tag 02 - Merchant Category Code (MCC)
- **Type**: String
- **Length**: 4 characters (fixed)
- **Encoding**: ASCII
- **Description**: Merchant category classification code

##### Tag 03 - Merchant Postal Code
- **Type**: String
- **Length**: Variable (up to 10 characters)
- **Encoding**: ASCII
- **Length Prefix**: LL (2-digit length indicator)
- **Description**: Merchant's postal/ZIP code

##### Tag 04 - Merchant Website
- **Type**: String
- **Length**: Variable (up to 299 characters)
- **Encoding**: ASCII
- **Length Prefix**: LLL (3-digit length indicator)
- **Description**: Merchant's website URL

### Field 11 - Systems Trace Audit Number (STAN)
- **Type**: String
- **Length**: 6 characters (fixed)
- **Encoding**: ASCII
- **Description**: Unique sequence number for transaction tracking

### Field 55 - Chip Data
- **Type**: Binary
- **Length**: Variable (up to 999 bytes)
- **Encoding**: Binary
- **Length Prefix**: LLL (3-digit length indicator)
- **Description**: Contains EMV chip data for card transactions

## Network Connection

### Transport Protocol
The client should establish a TCP connection to the server/issuer for message transmission.

### Message Framing
Each ISO 8583 message must be prefixed with a **message length header**:
- **Size**: 2 bytes
- **Encoding**: Big-endian binary
- **Content**: Length of the ISO 8583 message payload (excluding the header itself)

### Example
For a 45-byte ISO 8583 authorization request:
```
Request:  [0x00, 0x2D][45 bytes of ISO 8583 message]
Response: [0x00, 0x1F][31 bytes of ISO 8583 response]
```

The length header allows the receiver to know exactly how many bytes to read for the complete message, enabling proper message boundary detection over the TCP stream.
