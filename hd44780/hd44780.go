package hd44780

import (
	"fmt"
	"time"

	"github.com/d2r2/go-i2c"
)

// CustomChar represents the data for a custom character. Only bits 0 - 4 are used.
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

const (
	// Commands
	lcdClearDisplay byte = 0x01 // 00000001
	lcdReturnHome   byte = 0x02 // 00000010
	lcdCursorShift  byte = 0x10 // 00010000
	lcdSetCGRamAddr byte = 0x40 // 01000000
	lcdSetDDRamAddr byte = 0x80 // 10000000

	// Entry mode flags
	lcdEntryRight          byte = 0x00 // 00000000
	lcdEntryShiftDecrement byte = 0x00 // 00000000
	lcdEntryShiftIncrement byte = 0x01 // 00000001
	lcdEntryLeft           byte = 0x02 // 00000010
	lcdSetEntryMode        byte = 0x04 // 00000100

	// Display mode flags
	lcdDisplayOn      byte = 0x04 // 00000100
	lcdDisplayOff     byte = 0x00 // 00000000
	lcdCursorBlinkOn  byte = 0x01 // 00000001
	lcdCursorBlinkOff byte = 0x00 // 00000000
	lcdCursorOn       byte = 0x02 // 00000010
	lcdCursorOff      byte = 0x00 // 00000000
	lcdSetDisplayMode byte = 0x08 // 00001000

	// Cursor and display move flags
	lcdMoveLeft    byte = 0x00 // 00000000
	lcdCursorMove  byte = 0x00 // 00000000
	lcdMoveRight   byte = 0x04 // 00000100
	lcdDisplayMove byte = 0x08 // 00001000

	// Function mode flags
	lcd1Line           byte = 0x00 // 00000000
	lcd2Line           byte = 0x08 // 00001000
	lcd4BitMode        byte = 0x00 // 00000000
	lcd8BitMode        byte = 0x10 // 00010000
	lcd5x8Dots         byte = 0x00 // 00000000
	lcd5x10Dots        byte = 0x04 // 00000100
	lcdSetFunctionMode byte = 0x20 // 00100000

	// Backlight Control
	lcdBacklightOff byte = 0x00
	lcdBacklightOn  byte = 0x08

	// these appear to be different from https://github.com/davecheney/i2c/blob/master/helloworld/main.go#L15
	// same as https://github.com/d2r2/go-hd44780/blob/master/lcd.go#L41-L43
	// https://github.com/kidoman/embd/blob/master/controller/hd44780/hd44780.go#L571-L576
	// can't find the other pins for the places that use this format
	enableBit    byte = 0x4 // EN
	readWriteBit byte = 0x2 // RW

	registerSelectHigh byte = 0x1
	registerSelectLow  byte = 0x0
)

type Hd44780I2c struct {
	I2C       *i2c.I2C
	backlight bool
}

// NewHd44780I2c returns a new Connection based on an I²C bus.
func NewHd44780I2c(i2c *i2c.I2C) *Hd44780I2c {
	c := &Hd44780I2c{
		I2C:       i2c,
		backlight: true,
	}

	// initialisation
	time.Sleep(time.Millisecond * 20)
	c.writeByte(0x03, registerSelectLow) // init
	c.writeByte(0x03, registerSelectLow) // init
	c.writeByte(0x03, registerSelectLow) // init
	c.writeByte(0x02, registerSelectLow) // 4 bit mode

	c.writeByte(byte(lcdSetFunctionMode)|byte(lcd2Line)|byte(lcd5x8Dots)|byte(lcd4BitMode), registerSelectLow)
	c.writeByte(byte(lcdSetDisplayMode)|byte(lcdDisplayOn), registerSelectLow)
	c.writeByte(byte(lcdClearDisplay), registerSelectLow)
	c.writeByte(byte(lcdSetEntryMode|lcdEntryLeft), registerSelectLow)
	time.Sleep(time.Millisecond * 200)

	return c
}

func (this *Hd44780I2c) strobe(data byte) error {
	fmt.Println("in strobe")
	fmt.Printf("first write: %b\n", data|enableBit)
	_, err := this.I2C.WriteByte(data | enableBit)
	if err != nil {
		return err
	}
	time.Sleep(time.Microsecond * 200)
	_, err = this.I2C.WriteByte((data & ^enableBit))
	if err != nil {
		return err
	}
	time.Sleep(time.Microsecond * 30)
	return nil
}

func (this *Hd44780I2c) writeFourBits(data byte) error {
	fmt.Printf("write 4: %b\n", data)
	_, err := this.I2C.WriteByte(data)
	if err != nil {
		return err
	}
	err = this.strobe(data)
	if err != nil {
		return err
	}
	return nil
}

func (this *Hd44780I2c) prepareFirstFourBits(data byte) byte {
	if this.backlight {
		return (data & 0xF0) | lcdBacklightOn
	} else {
		return (data & 0xF0) | lcdBacklightOff
	}
}

func (this *Hd44780I2c) prepareSecondFourBits(data byte) byte {
	if this.backlight {
		return ((data << 4) & 0xF0) | lcdBacklightOn
	} else {
		return ((data << 4) & 0xF0) | lcdBacklightOff
	}
}

func (this *Hd44780I2c) writeByte(data, rs byte) error {
	first := this.prepareFirstFourBits(data)
	second := this.prepareSecondFourBits(data)

	// first 4 bits
	err := this.writeFourBits(byte(rs) | first)
	if err != nil {
		return err
	}

	// second 4 bits
	err = this.writeFourBits(byte(rs) | second)
	if err != nil {
		return err
	}
	return err
}

func (this *Hd44780I2c) DisplayString(str string, line, pos byte) error {
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

	err := this.writeByte(0x80+address, registerSelectLow)
	if err != nil {
		return err
	}
	for _, c := range str {
		err = this.writeByte(byte(c), registerSelectHigh)
		if err != nil {
			return err
		}
	}
	return nil
}

// Clear lcd and set position to home
func (this *Hd44780I2c) Clear() error {
	err := this.writeByte(lcdClearDisplay, registerSelectLow)
	if err != nil {
		return err
	}
	err = this.writeByte(lcdReturnHome, registerSelectLow)
	if err != nil {
		return err
	}
	return nil
}

func (this *Hd44780I2c) BacklightOn() error {
	this.backlight = true
	_, err := this.I2C.WriteByte(lcdBacklightOn)
	return err
}

func (this *Hd44780I2c) BacklightOff() error {
	this.backlight = false
	_, err := this.I2C.WriteByte(lcdBacklightOff)
	return err
}

func (this *Hd44780I2c) LoadCustomChars(chars [8]CustomChar) error {
	err := this.writeByte(lcdSetCGRamAddr, registerSelectLow)
	if err != nil {
		return err
	}

	for _, c := range chars {
		for _, b := range c {
			err = this.writeByte(b, registerSelectHigh)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
