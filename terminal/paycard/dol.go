package paycard

import (
	"bytes"
	"errors"
	"fmt"
)

// DOL (Data Object List) represents a parsed tag, its length
type DOL struct {
	Tag    string // Tag is a byte slice to support multi-byte tags
	Length int
}

// ParseDOL parses a PDOL byte stream into a slice of TagAndLength with index tracking
func ParseDOL(data []byte) ([]DOL, error) {
	var dolList []DOL

	if data == nil || len(data) == 0 {
		return nil, errors.New("input data is empty or nil")
	}

	stream := bytes.NewReader(data)

	for stream.Len() > 0 {
		// Read the tag, which could be 1 or more bytes
		tag, err := readTag(stream)
		if err != nil {
			return nil, fmt.Errorf("error reading tag: %v", err)
		}

		// Read the length, which is typically 1 byte but could vary in specific formats
		lengthByte := make([]byte, 1)
		if _, err := stream.Read(lengthByte); err != nil {
			return nil, fmt.Errorf("error reading length: %v", err)
		}
		length := int(lengthByte[0])
		tagName := fmt.Sprintf("%X", tag)

		// Append the tag, length
		dolList = append(dolList, DOL{
			Tag:    tagName,
			Length: length,
		})
	}

	return dolList, nil
}

// readTag reads a tag from the byte stream, handling multi-byte tags
func readTag(stream *bytes.Reader) ([]byte, error) {
	tag := make([]byte, 1)
	if _, err := stream.Read(tag); err != nil {
		return nil, err
	}

	// Check if the tag is multi-byte (0x1F or higher indicates multi-byte tag)
	if tag[0]&0x1F == 0x1F {
		for {
			nextByte := make([]byte, 1)
			if _, err := stream.Read(nextByte); err != nil {
				return nil, err
			}
			tag = append(tag, nextByte[0])

			// If the most significant bit (MSB) is not set, this is the last byte of the tag
			if nextByte[0]&0x80 == 0 {
				break
			}
		}
	}

	return tag, nil
}

// PrettyPrintDOL prints a list of TagAndLength in a human-readable format
func PrettyPrintDOL(tagAndLengthList []DOL) {

	for _, tl := range tagAndLengthList {
		tagName, found := GetPDOLTagInfo(tl.Tag)
		if !found {
			tagName = PDOLTagInfo{Description: "Unknown", Definition: "Unknown"}
		}
		fmt.Printf("Tag: %s \tLength: %d \t%s\t%s\n", tl.Tag, tl.Length, tagName.Definition, tagName.Description)
	}
}
