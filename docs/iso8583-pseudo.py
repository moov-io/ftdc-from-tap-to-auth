// Step 1: Prepare field data
fields = {
    2: "4242424242424242",           // PAN
    3: "000150",                     // Amount ($1.50)
    4: "20250728143022000000",       // Transmission DateTime
    7: "840",                        // Currency (USD)
    11: "123456"                     // STAN
}

// Step 2: Pack data fields and build bitmap
data = ""
bitmap = new_bitmap()

for field_number, field_value in fields:

    // Encode field according to spec
    encoded_field = spec[field_number].encode(field_value)
    data = data + encoded_field
    
    // Mark field as present in bitmap
    bitmap.set_bit(field_number)

// Step 3: Assemble final message
mti = "0100"                         // Message Type Indicator

message = mti + bitmap + data
