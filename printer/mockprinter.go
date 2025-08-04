package printer

import (
	"fmt"
)

type MockPrinter struct {
}

func NewMockPrinter() (*MockPrinter, error) {
	printer := &MockPrinter{}
	return printer, nil
}

func (p *MockPrinter) sendCommand(data []byte) error {
	// Convert ESC commands to human-readable format for mock output
	output := p.convertEscCommands(data)
	fmt.Printf("[MOCK PRINTER] %s\n", output)
	return nil
}

func (p *MockPrinter) convertEscCommands(data []byte) string {
	// Convert raw ESC commands to human-readable format
	result := ""
	i := 0
	for i < len(data) {
		if i+1 < len(data) && data[i] == 0x1B {
			// ESC sequence
			if i+2 < len(data) {
				switch data[i+1] {
				case 0x40:
					result += "[INIT]"
					i += 2
				case 0x61:
					if i+3 <= len(data) {
						switch data[i+2] {
						case 0x00:
							result += "[ALIGN_LEFT]"
						case 0x01:
							result += "[ALIGN_CENTER]"
						case 0x02:
							result += "[ALIGN_RIGHT]"
						default:
							result += fmt.Sprintf("[ALIGN_UNKNOWN_%02X]", data[i+2])
						}
						i += 3
					} else {
						result += "[ALIGN_INCOMPLETE]"
						i += 2
					}
				case 0x45:
					if i+3 < len(data) {
						if data[i+2] == 0x01 {
							result += "[BOLD_ON]"
						} else if data[i+2] == 0x00 {
							result += "[BOLD_OFF]"
						} else {
							result += fmt.Sprintf("[BOLD_UNKNOWN_%02X]", data[i+2])
						}
						i += 3
					} else {
						result += "[BOLD_INCOMPLETE]"
						i += 2
					}
				default:
					result += fmt.Sprintf("[ESC_%02X]", data[i+1])
					i += 2
				}
			} else {
				result += "[ESC_INCOMPLETE]"
				i++
			}
		} else if i < len(data) && data[i] == 0x1D {
			// GS sequence (cut command)
			if i+4 <= len(data) && data[i+1] == 0x56 && data[i+2] == 0x42 && data[i+3] == 0x00 {
				result += "[CUT_PAPER]"
				i += 4
			} else if i+1 < len(data) {
				result += fmt.Sprintf("[GS_%02X]", data[i+1])
				i += 2
			} else {
				result += "[GS_INCOMPLETE]"
				i++
			}
		} else if i < len(data) && data[i] == 0x0A {
			// Line feed
			result += "[LF]"
			i++
		} else if i < len(data) {
			// Regular text
			result += string(data[i])
			i++
		}
	}
	return result
}

func (p *MockPrinter) PrintText(text string) error {
	// Send text as bytes
	return p.sendCommand([]byte(text))
}

func (p *MockPrinter) PrintLine(text string) error {
	// Print text with line feed
	data := append([]byte(text), ESC_FEED...)
	return p.sendCommand(data)
}

func (p *MockPrinter) PrintBold(text string) error {
	// Bold on, text, bold off, line feed
	var data []byte
	data = append(data, ESC_BOLD_ON...)
	data = append(data, []byte(text)...)
	data = append(data, ESC_BOLD_OFF...)
	data = append(data, ESC_FEED...)
	return p.sendCommand(data)
}

func (p *MockPrinter) PrintTitle(title string) error {
	// Center align, bold on, title, bold off, line feed, left align
	var data []byte
	data = append(data, ESC_ALIGN_C...)
	data = append(data, ESC_BOLD_ON...)
	data = append(data, []byte(title)...)
	data = append(data, ESC_BOLD_OFF...)
	data = append(data, ESC_FEED...)
	data = append(data, ESC_ALIGN_L...)
	return p.sendCommand(data)
}

func (p *MockPrinter) PrintCentered(text string) error {
	// Center align, text, left align, line feed
	var data []byte
	data = append(data, ESC_ALIGN_C...)
	data = append(data, []byte(text)...)
	data = append(data, ESC_FEED...)
	data = append(data, ESC_ALIGN_L...)
	return p.sendCommand(data)
}

func (p *MockPrinter) Feed(lines int) error {
	for range lines {
		if err := p.sendCommand(ESC_FEED); err != nil {
			return err
		}
	}
	return nil
}

func (p *MockPrinter) Cut() error {
	return p.sendCommand(ESC_CUT)
}

func (p *MockPrinter) PrintLines(lines []string) error {
	for _, line := range lines {
		// cut off the line if it exceeds 32 characters
		if len(line) > 32 {
			line = line[:32]
		}

		if err := p.PrintLine(line); err != nil {
			return fmt.Errorf("failed to print line '%s': %v", line, err)
		}
	}
	return nil
}

func (p *MockPrinter) PrintBitmapImage(bitmap *BitmapImage) error {
	p.PrintLine("IMAGE PLACEHOLDER")
	return nil
}

func (p *MockPrinter) Close() error {
	return nil
}
