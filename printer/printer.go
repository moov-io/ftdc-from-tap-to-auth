package printer

import (
	"fmt"

	"github.com/google/gousb"
)

const (
	// Your printer's USB identifiers
	VendorID  = 0x0fe6 // 4070
	ProductID = 0x811e // 33054
)

// ESC/POS commands for thermal printers
var (
	ESC_INIT     = []byte{0x1B, 0x40}             // Initialize printer
	ESC_ALIGN_L  = []byte{0x1B, 0x61, 0x00}       // Left align
	ESC_ALIGN_C  = []byte{0x1B, 0x61, 0x01}       // Center align
	ESC_ALIGN_R  = []byte{0x1B, 0x61, 0x02}       // Right align
	ESC_BOLD_ON  = []byte{0x1B, 0x45, 0x01}       // Bold on
	ESC_BOLD_OFF = []byte{0x1B, 0x45, 0x00}       // Bold off
	ESC_FEED     = []byte{0x0A}                   // Line feed
	ESC_CUT      = []byte{0x1D, 0x56, 0x42, 0x00} // Cut paper
)

type Printer interface {
	PrintText(text string) error
	PrintLine(text string) error
	PrintBold(text string) error
	PrintTitle(title string) error
	PrintCentered(text string) error
	Feed(lines int) error
	Cut() error
	PrintLines(lines []string) error
	Close() error
}

type ThermalPrinter struct {
	ctx    *gousb.Context
	device *gousb.Device
	intf   *gousb.Interface
	out    *gousb.OutEndpoint
}

func NewThermalPrinter() (*ThermalPrinter, error) {
	ctx := gousb.NewContext()

	// Find the printer device
	device, err := ctx.OpenDeviceWithVIDPID(VendorID, ProductID)
	if err != nil {
		ctx.Close()
		return nil, fmt.Errorf("failed to open device: %v", err)
	}

	if device == nil {
		ctx.Close()
		return nil, fmt.Errorf("printer device not found (VID: 0x%04x, PID: 0x%04x)", VendorID, ProductID)
	}

	// Get device configuration
	config, err := device.Config(1)
	if err != nil {
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("failed to get config: %v", err)
	}

	// Claim interface (usually interface 0 for printers)
	intf, err := config.Interface(0, 0)
	if err != nil {
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("failed to claim interface: %v", err)
	}

	// Find bulk out endpoint
	var outEndpoint *gousb.OutEndpoint
	for _, endpoint := range intf.Setting.Endpoints {
		if endpoint.Direction == gousb.EndpointDirectionOut &&
			endpoint.TransferType == gousb.TransferTypeBulk {
			outEndpoint, err = intf.OutEndpoint(endpoint.Number)
			if err != nil {
				intf.Close()
				device.Close()
				ctx.Close()
				return nil, fmt.Errorf("failed to get out endpoint: %v", err)
			}
			break
		}
	}

	if outEndpoint == nil {
		intf.Close()
		device.Close()
		ctx.Close()
		return nil, fmt.Errorf("no bulk out endpoint found")
	}

	printer := &ThermalPrinter{
		ctx:    ctx,
		device: device,
		intf:   intf,
		out:    outEndpoint,
	}

	// Initialize printer
	if err := printer.sendCommand(ESC_INIT); err != nil {
		printer.Close()
		return nil, fmt.Errorf("failed to initialize printer: %v", err)
	}

	return printer, nil
}

func (tp *ThermalPrinter) sendCommand(data []byte) error {
	_, err := tp.out.Write(data)
	return err
}

func (tp *ThermalPrinter) PrintText(text string) error {
	// Send text as bytes
	return tp.sendCommand([]byte(text))
}

func (tp *ThermalPrinter) PrintLine(text string) error {
	// Print text with line feed
	data := append([]byte(text), ESC_FEED...)
	return tp.sendCommand(data)
}

func (tp *ThermalPrinter) PrintBold(text string) error {
	// Bold on, text, bold off, line feed
	var data []byte
	data = append(data, ESC_BOLD_ON...)
	data = append(data, []byte(text)...)
	data = append(data, ESC_BOLD_OFF...)
	data = append(data, ESC_FEED...)
	return tp.sendCommand(data)
}

func (tp *ThermalPrinter) PrintTitle(title string) error {
	// Center align, bold on, title, bold off, line feed, left align
	var data []byte
	data = append(data, ESC_ALIGN_C...)
	data = append(data, ESC_BOLD_ON...)
	data = append(data, []byte(title)...)
	data = append(data, ESC_BOLD_OFF...)
	data = append(data, ESC_FEED...)
	data = append(data, ESC_ALIGN_L...)
	return tp.sendCommand(data)
}

func (tp *ThermalPrinter) PrintCentered(text string) error {
	// Center align, text, left align, line feed
	var data []byte
	data = append(data, ESC_ALIGN_C...)
	data = append(data, []byte(text)...)
	data = append(data, ESC_FEED...)
	data = append(data, ESC_ALIGN_L...)
	return tp.sendCommand(data)
}

func (tp *ThermalPrinter) Feed(lines int) error {
	for i := 0; i < lines; i++ {
		if err := tp.sendCommand(ESC_FEED); err != nil {
			return err
		}
	}
	return nil
}

func (tp *ThermalPrinter) Cut() error {
	return tp.sendCommand(ESC_CUT)
}

func (tp *ThermalPrinter) PrintLines(lines []string) error {
	for _, line := range lines {
		// cut off the line if it exceeds 32 characters
		if len(line) > 32 {
			line = line[:32]
		}

		if err := tp.PrintLine(line); err != nil {
			return fmt.Errorf("failed to print line '%s': %v", line, err)
		}
	}
	return nil
}

func (tp *ThermalPrinter) Close() error {
	if tp.intf != nil {
		tp.intf.Close()
	}
	if tp.device != nil {
		tp.device.Close()
	}
	if tp.ctx != nil {
		tp.ctx.Close()
	}
	return nil
}
