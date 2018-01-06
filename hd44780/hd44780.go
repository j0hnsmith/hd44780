package hd44780

import (
	"fmt"
	"time"

	"github.com/d2r2/go-i2c"
)

type entryMode byte
type displayMode byte
type functionMode byte
type registerSelect byte

// CustomChar represents the data for a custom character. Only bits 0 - 4 are used (least significant).
// Index 0 is the topmost line, bits set to 1 are 'on'. There's a nice generator that outputs hex
// https://www.quinapalus.com/hd44780udg.html
//
// Here's an example of 7 varying states of a battery charge
//		chars := [8]hd44780.CustomChar{
//			hd44780.CustomChar{0xe, 0x1b, 0x11, 0x11, 0x11, 0x11, 0x11, 0x1f},
//			hd44780.CustomChar{0xe, 0x1b, 0x11, 0x11, 0x11, 0x11, 0x1f, 0x1f},
//			hd44780.CustomChar{0xe, 0x1b, 0x11, 0x11, 0x11, 0x1f, 0x1f, 0x1f},
//			hd44780.CustomChar{0xe, 0x1b, 0x11, 0x11, 0x1f, 0x1f, 0x1f, 0x1f},
//			hd44780.CustomChar{0xe, 0x1b, 0x11, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f},
//			hd44780.CustomChar{0xe, 0x1b, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f},
//			hd44780.CustomChar{0xe, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f},
//			hd44780.CustomChar{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
//		}
type CustomChar [8]byte

// BacklightPolarity is used to set the polarity of the backlight switch, either positive or negative.
type BacklightPolarity bool

const (
	// Negative indicates that the backlight is active-low and must have a logical low value to enable.
	Negative BacklightPolarity = false
	// Positive indicates that the backlight is active-high and must have a logical high value to enable.
	Positive BacklightPolarity = true

	// delay values from http://irtfweb.ifa.hawaii.edu/~tcs3/tcs3/vendor_info/Technologic_systems/embeddedx86/HD44780_LCD/lcd0.shtml.htm#instruction_set
	writeDelay = 40 * time.Microsecond
	pulseDelay = 1 * time.Microsecond
	clearDelay = 1640 * time.Microsecond

	// delays from datasheet https://www.sparkfun.com/datasheets/LCD/HD44780.pdf
	initDelay1 = 4100 * time.Microsecond
	initDelay2 = 100 * time.Microsecond

	// Commands
	lcdClearDisplay byte = 0x01 // 00000001
	lcdReturnHome   byte = 0x02 // 00000010
	lcdCursorShift  byte = 0x10 // 00010000
	lcdSetCGRamAddr byte = 0x40 // 01000000
	lcdSetDDRamAddr byte = 0x80 // 10000000

	// Entry mode flags
	lcdSetEntryMode   entryMode = 0x04 // 00000100
	lcdEntryDecrement entryMode = 0x00 // 00000000
	lcdEntryIncrement entryMode = 0x02 // 00000010
	lcdEntryShiftOff  entryMode = 0x00 // 00000000
	lcdEntryShiftOn   entryMode = 0x01 // 00000001

	// Display mode flags
	lcdDisplayOn          displayMode = 0x04 // 00000100
	lcdDisplayOff         displayMode = 0x00 // 00000000
	lcdBlinkCursorOn      displayMode = 0x01 // 00000001
	lcdBlinkCursorOff     displayMode = 0x00 // 00000000
	lcdUnderlineCursorOn  displayMode = 0x02 // 00000010
	lcdUnderlineCursorOff displayMode = 0x00 // 00000000
	lcdSetDisplayMode     displayMode = 0x08 // 00001000

	// Cursor and display move flags
	lcdMoveLeft    byte = 0x00 // 00000000
	lcdCursorMove  byte = 0x00 // 00000000
	lcdMoveRight   byte = 0x04 // 00000100
	lcdDisplayMove byte = 0x08 // 00001000

	// Function mode flags
	lcd1Line           functionMode = 0x00 // 00000000
	lcd2Line           functionMode = 0x08 // 00001000
	lcd4BitMode        functionMode = 0x00 // 00000000
	lcd8BitMode        functionMode = 0x10 // 00010000
	lcd5x8Dots         functionMode = 0x00 // 00000000
	lcd5x10Dots        functionMode = 0x04 // 00000100
	lcdSetFunctionMode functionMode = 0x20 // 00100000

	// Backlight Control
	lcdBacklightOff byte = 0x00
	lcdBacklightOn  byte = 0x08

	// these appear to be different from https://github.com/davecheney/i2c/blob/master/helloworld/main.go#L15
	// same as https://github.com/d2r2/go-hd44780/blob/master/lcd.go#L41-L43
	// https://github.com/kidoman/embd/blob/master/controller/hd44780/hd44780.go#L571-L576
	// can't find the other pins for the places that use this format
	enableBit    byte = 0x4 // EN
	readWriteBit byte = 0x2 // RW

	registerSelectHigh registerSelect = 0x1
	registerSelectLow  registerSelect = 0x0
// I2CPinMap represents a mapping between the pins on an I²C port expander and
// the pins on the HD44780 controller.
type I2CPinMap struct {
	RS, RW, EN     byte
	D4, D5, D6, D7 byte
	Backlight      byte
	BLPolarity     BacklightPolarity
}

var (
	// MJKDZPinMap is the standard pin mapping for an MJKDZ-based I²C backpack.
	MJKDZPinMap I2CPinMap = I2CPinMap{
		RS: 6, RW: 5, EN: 4,
		D4: 0, D5: 1, D6: 2, D7: 3,
		Backlight:  7,
		BLPolarity: Negative,
	}
	// PCF8574PinMap is the standard pin mapping for a PCF8574-based I²C backpack.
	PCF8574PinMap I2CPinMap = I2CPinMap{
		RS: 0, RW: 1, EN: 2,
		D4: 4, D5: 5, D6: 6, D7: 7,
		Backlight:  3,
		BLPolarity: Positive,
	}
)

type Hd44780I2c struct {
	I2C       *i2c.I2C
	PinMap    I2CPinMap
	backlight bool
	eMode     entryMode
	dMode     displayMode
	fMode     functionMode
}

// NewHd44780I2c returns a new Connection based on an I²C bus.
func NewHd44780I2c(i2c *i2c.I2C, pinMap I2CPinMap, modes ...ModeSetter) (*Hd44780I2c, error) {
	c := &Hd44780I2c{
		I2C:       i2c,
		PinMap:    pinMap,
		backlight: true,
		eMode:     0x00,
		dMode:     0x00,
		fMode:     0x00,
	}

	err := c.lcdInit()
	if err != nil {
		return nil, err
	}

	err = c.SetMode(append(DefaultModes, modes...)...)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (hd *Hd44780I2c) lcdInit() error {
	time.Sleep(time.Millisecond * 20)
	err := hd.WriteInstruction(0x03) // init
	if err != nil {
		return err
	}

	time.Sleep(initDelay1)

	err = hd.WriteInstruction(0x03) // init
	if err != nil {
		return err
	}

	time.Sleep(initDelay2)

	err = hd.WriteInstruction(0x03) // init
	if err != nil {
		return err
	}

	err = hd.WriteInstruction(0x02) // 4 bit mode
	if err != nil {
		return err
	}

	return hd.Clear()
}

// SetModes modifies the entry mode, display mode, and function mode with the
// given mode setter functions.
func (hd *Hd44780I2c) SetMode(modes ...ModeSetter) error {
	for _, m := range modes {
		m(hd)
	}
	functions := []func() error{
		func() error { return hd.setEntryMode() },
		func() error { return hd.setDisplayMode() },
		func() error { return hd.setFunctionMode() },
	}
	for _, f := range functions {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

func (hd *Hd44780I2c) setEntryMode() error {
	return hd.WriteInstruction(byte(lcdSetEntryMode | hd.eMode))
}

func (hd *Hd44780I2c) setDisplayMode() error {
	return hd.WriteInstruction(byte(lcdSetDisplayMode | hd.dMode))
}

func (hd *Hd44780I2c) setFunctionMode() error {
	return hd.WriteInstruction(byte(lcdSetFunctionMode | hd.fMode))
}

// write writes a register select flag and byte to the I²C connection.
func (hd *Hd44780I2c) write(data byte, rs registerSelect) error {
	var instructionHigh byte = 0x00
	instructionHigh |= ((data >> 4) & 0x01) << hd.PinMap.D4
	instructionHigh |= ((data >> 5) & 0x01) << hd.PinMap.D5
	instructionHigh |= ((data >> 6) & 0x01) << hd.PinMap.D6
	instructionHigh |= ((data >> 7) & 0x01) << hd.PinMap.D7

	var instructionLow byte = 0x00
	instructionLow |= ((data >> 0) & 0x01) << hd.PinMap.D4
	instructionLow |= ((data >> 1) & 0x01) << hd.PinMap.D5
	instructionLow |= ((data >> 2) & 0x01) << hd.PinMap.D6
	instructionLow |= ((data >> 3) & 0x01) << hd.PinMap.D7

	instructions := []byte{instructionHigh, instructionLow}
	for _, ins := range instructions {
		ins |= byte(rs) << hd.PinMap.RS
		if hd.backlight == bool(hd.PinMap.BLPolarity) {
			ins |= 0x01 << hd.PinMap.Backlight
		}

		bytes := []byte{ins, ins | (0x01 << hd.PinMap.EN), ins}
		for _, b := range bytes {
			time.Sleep(pulseDelay)
			_, err := hd.I2C.WriteByte(b)
			if err != nil {
				return err
			}
		}
	}
	time.Sleep(writeDelay) // is this necessary with i2c?
	return nil
}

func (hd *Hd44780I2c) DisplayString(str string, line, pos byte) error {
	var address byte
	switch line {
	case 1:
		address = pos
	case 2:
		address = 0x40 + pos
	case 3:
		address = 0x10 + pos
	case 4:
		address = 0x54 + pos
	}

	err := hd.WriteInstruction(lcdSetDDRamAddr + address)
	if err != nil {
		return err
	}
	for _, c := range str {
		err = hd.WriteChar(byte(c))
		if err != nil {
			return err
		}
	}
	return nil
}

func (hd *Hd44780I2c) Write(buf []byte) (int, error) {
	for i, c := range buf {
		err := hd.WriteChar(c)
		if err != nil {
			return maxInt(i-1, 0), err
		}
	}
	return len(buf), nil
}

// SetDDRamAddr sets the input cursor to the given address.
func (hd *Hd44780I2c) SetDDRamAddr(value byte) error {
	return hd.WriteInstruction(lcdSetDDRamAddr | value)
}

// WriteChar writes a byte to the bus with register select in data mode.
func (hd *Hd44780I2c) WriteChar(value byte) error {
	return hd.write(value, registerSelectHigh)
}

// WriteInstruction writes a byte to the bus with register select in command mode.
func (hd *Hd44780I2c) WriteInstruction(value byte) error {
	return hd.write(value, registerSelectLow)
}

func (hd *Hd44780I2c) BacklightOn() error {
	hd.backlight = true
	_, err := hd.I2C.WriteByte(lcdBacklightOn)
	return err
}

func (hd *Hd44780I2c) BacklightOff() error {
	hd.backlight = false
	_, err := hd.I2C.WriteByte(lcdBacklightOff)
	return err
}

// DisplayOff sets the display mode to off.
func (hd *Hd44780I2c) DisplayOff() error {
	DisplayOff(hd)
	return hd.setDisplayMode()
}

// DisplayOn sets the display mode to on.
func (hd *Hd44780I2c) DisplayOn() error {
	DisplayOn(hd)
	return hd.setDisplayMode()
}

// UnderlineCursorOff turns the cursor off.
func (hd *Hd44780I2c) UnderlineCursorOff() error {
	UnderlineCursorOff(hd)
	return hd.setDisplayMode()
}

// UnderlineCursorOn turns the cursor on.
func (hd *Hd44780I2c) UnderlineCursorOn() error {
	UnderlineCursorOn(hd)
	return hd.setDisplayMode()
}

// BlinkCursorOff sets cursor blink mode off.
func (hd *Hd44780I2c) BlinkCursorOff() error {
	BlinkCursorOff(hd)
	return hd.setDisplayMode()
}

// BlinkCursorOn sets cursor blink mode on.
func (hd *Hd44780I2c) BlinkCursorOn() error {
	BlinkCursorOn(hd)
	return hd.setDisplayMode()
}

// EntryShiftOn sets entry shift on, moves all the text one space each time a letter is added.
func (hd *Hd44780I2c) EntryShiftOn() error {
	EntryShiftOn(hd)
	return hd.setEntryMode()
}

// EntryShiftOn sets entry shift off.
func (hd *Hd44780I2c) EntryShiftOff() error {
	EntryShiftOff(hd)
	return hd.setEntryMode()
}

// ShiftLeft shifts the cursor and all characters to the left.
func (hd *Hd44780I2c) ShiftLeft() error {
	return hd.WriteInstruction(lcdCursorShift | lcdDisplayMove | lcdMoveLeft)
}

// ShiftRight shifts the cursor and all characters to the right.
func (hd *Hd44780I2c) ShiftRight() error {
	return hd.WriteInstruction(lcdCursorShift | lcdDisplayMove | lcdMoveRight)
}

// Home moves the cursor and all characters to the home position.
func (hd *Hd44780I2c) Home() error {
	err := hd.WriteInstruction(lcdReturnHome)
	return err
}

// Clear clears the display and mode settings sets the cursor to the home position.
func (hd *Hd44780I2c) Clear() error {
	err := hd.WriteInstruction(lcdClearDisplay)
	if err != nil {
		return err
	}
	time.Sleep(clearDelay)
	// have to set mode here because clear also clears some mode settings
	return hd.SetMode()
}

// LoadCustomChars stores 8 custom characters into CGRAM, see type CustomChar docs for an example.
func (hd *Hd44780I2c) LoadCustomChars(chars [8]CustomChar) error {
	err := hd.WriteInstruction(lcdSetCGRamAddr)
	if err != nil {
		return err
	}

	for _, c := range chars {
		for _, b := range c {
			err = hd.WriteChar(b)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DefaultModes are the default initialization modes for an HD44780.
// ModeSetters passed in to a constructor will override these default values.
var DefaultModes []ModeSetter = []ModeSetter{
	FourBitMode,
	TwoLine,
	Dots5x8,
	EntryIncrement,
	EntryShiftOff,
	DisplayOn,
	UnderlineCursorOff,
	BlinkCursorOff,
}

// ModeSetter defines a function used for setting modes on an HD44780.
// ModeSetters must be used with the SetMode function or in a constructor.
type ModeSetter func(*Hd44780I2c)

// EntryDecrement is a ModeSetter that sets the HD44780 to entry decrement mode.
func EntryDecrement(hd *Hd44780I2c) { hd.eMode &= ^lcdEntryIncrement }

// EntryIncrement is a ModeSetter that sets the HD44780 to entry increment mode.
func EntryIncrement(hd *Hd44780I2c) { hd.eMode |= lcdEntryIncrement }

// EntryShiftOff is a ModeSetter that sets the HD44780 to entry shift off mode.
func EntryShiftOff(hd *Hd44780I2c) { hd.eMode &= ^lcdEntryShiftOn }

// EntryShiftOn is a ModeSetter that sets the HD44780 to entry shift on mode.
func EntryShiftOn(hd *Hd44780I2c) { hd.eMode |= lcdEntryShiftOn }

// DisplayOff is a ModeSetter that sets the HD44780 to display off mode.
func DisplayOff(hd *Hd44780I2c) { hd.dMode &= ^lcdDisplayOn }

// DisplayOn is a ModeSetter that sets the HD44780 to display on mode.
func DisplayOn(hd *Hd44780I2c) { hd.dMode |= lcdDisplayOn }

// UnderlineCursorOff is a ModeSetter that sets the HD44780 to cursor off mode.
func UnderlineCursorOff(hd *Hd44780I2c) { hd.dMode &= ^lcdUnderlineCursorOn }

// UnderlineCursorOn is a ModeSetter that sets the HD44780 to cursor on mode.
func UnderlineCursorOn(hd *Hd44780I2c) { hd.dMode |= lcdUnderlineCursorOn }

// BlinkCursorOff is a ModeSetter that sets the HD44780 to cursor blink off mode.
func BlinkCursorOff(hd *Hd44780I2c) { hd.dMode &= ^lcdBlinkCursorOn }

// BlinkCursorOn is a ModeSetter that sets the HD44780 to cursor blink on mode.
func BlinkCursorOn(hd *Hd44780I2c) { hd.dMode |= lcdBlinkCursorOn }

// FourBitMode is a ModeSetter that sets the HD44780 to 4-bit bus mode.
func FourBitMode(hd *Hd44780I2c) { hd.fMode &= ^lcd8BitMode }

// EightBitMode is a ModeSetter that sets the HD44780 to 8-bit bus mode.
func EightBitMode(hd *Hd44780I2c) { hd.fMode |= lcd8BitMode }

// OneLine is a ModeSetter that sets the HD44780 to 1-line display mode.
func OneLine(hd *Hd44780I2c) { hd.fMode &= ^lcd2Line }

// TwoLine is a ModeSetter that sets the HD44780 to 2-line display mode.
func TwoLine(hd *Hd44780I2c) { hd.fMode |= lcd2Line }

// Dots5x8 is a ModeSetter that sets the HD44780 to 5x8-pixel character mode.
func Dots5x8(hd *Hd44780I2c) { hd.fMode &= ^lcd5x10Dots }

// Dots5x10 is a ModeSetter that sets the HD44780 to 5x10-pixel character mode.
func Dots5x10(hd *Hd44780I2c) { hd.fMode |= lcd5x10Dots }

// EntryIncrementEnabled returns true if entry increment mode is enabled.
func (hd *Hd44780I2c) EntryIncrementEnabled() bool { return hd.eMode&lcdEntryIncrement > 0 }

// EntryShiftEnabled returns true if entry shift mode is enabled.
func (hd *Hd44780I2c) EntryShiftEnabled() bool { return hd.eMode&lcdEntryShiftOn > 0 }

// DisplayEnabled returns true if the display is on.
func (hd *Hd44780I2c) DisplayEnabled() bool { return hd.dMode&lcdDisplayOn > 0 }

// CursorEnabled returns true if the cursor is on.
func (hd *Hd44780I2c) CursorEnabled() bool { return hd.dMode&lcdUnderlineCursorOn > 0 }

// BlinkEnabled returns true if the cursor blink mode is enabled.
func (hd *Hd44780I2c) BlinkEnabled() bool { return hd.dMode&lcdBlinkCursorOn > 0 }

// EightBitModeEnabled returns true if 8-bit bus mode is enabled and false if 4-bit
// bus mode is enabled.
func (hd *Hd44780I2c) EightBitModeEnabled() bool { return hd.fMode&lcd8BitMode > 0 }

// TwoLineEnabled returns true if 2-line display mode is enabled and false if 1-line
// display mode is enabled.
func (hd *Hd44780I2c) TwoLineEnabled() bool { return hd.fMode&lcd2Line > 0 }

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
