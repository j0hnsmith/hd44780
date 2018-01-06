package hd44780_test

import (
	"fmt"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/j0hnsmith/hd44780"
)

func Example() {
	// errors are ignored in this example

	conn, _ := i2c.NewI2C(0x3F, 1) // i2cdetect to find address
	lcd, _ := hd44780.NewHd44780I2c(
		conn,
		hd44780.PCF8574PinMap,
		hd44780.RowAddress20Col,
		hd44780.UnderlineCursorOn,
	)

	lcd.DisplayString("Backlight off", 0, 0)
	time.Sleep(time.Second)
	lcd.BacklightOff()
	time.Sleep(time.Second)
	lcd.BacklightOn()
	lcd.DisplayString("Backlight on ", 0, 0)

	time.Sleep(time.Second)
	lcd.Clear()
	lcd.UnderlineCursorOn()
	time.Sleep(time.Second)
	lcd.DisplayString("Underline cursor", 0, 0)
	time.Sleep(time.Second * 2)
	lcd.UnderlineCursorOff()
	time.Sleep(time.Second)
	lcd.Clear()
	lcd.DisplayString("Blinking cursor", 0, 0)
	lcd.BlinkCursorOn()
	time.Sleep(time.Second * 2)
	lcd.BlinkCursorOff()

	lcd.Clear()
	lcd.DisplayString("Test display off/on", 0, 0)
	time.Sleep(time.Second)

	lcd.DisplayOff()
	time.Sleep(time.Second)
	lcd.DisplayOn()

	lcd.Clear()
	lcd.DisplayString("line 1", 0, 0)
	lcd.DisplayString("line 2", 1, 0)
	//lcd.DisplayString("line 3", 2, 0)
	//lcd.DisplayString("line 4", 3, 0)
	time.Sleep(time.Second)
	lcd.Clear()

	s := "write chars"
	for _, c := range s {
		lcd.WriteChar(byte(c))
		time.Sleep(time.Millisecond * 500)
	}
	time.Sleep(time.Second)
	lcd.Clear()

	lcd.DisplayString(fmt.Sprintf("all chars, 30%bC", 223), 0, 0)
	time.Sleep(time.Second)
	lcd.Clear()

	// load up to 8 of your own characters
	chars := [8]hd44780.CustomChar{
		hd44780.CustomChar{0xe, 0x1b, 0x11, 0x11, 0x11, 0x11, 0x11, 0x1f},
		hd44780.CustomChar{0xe, 0x1b, 0x11, 0x11, 0x11, 0x11, 0x1f, 0x1f},
		hd44780.CustomChar{0xe, 0x1b, 0x11, 0x11, 0x11, 0x1f, 0x1f, 0x1f},
		hd44780.CustomChar{0xe, 0x1b, 0x11, 0x11, 0x1f, 0x1f, 0x1f, 0x1f},
		hd44780.CustomChar{0xe, 0x1b, 0x11, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f},
		hd44780.CustomChar{0xe, 0x1b, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f},
		hd44780.CustomChar{0xe, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f},
		hd44780.CustomChar{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	}
	lcd.LoadCustomChars(chars)

	lcd.Clear()
	lcd.Home()

	lcd.DisplayString(fmt.Sprintf("%c%c%c%c%c%c%c", 0, 1, 2, 3, 4, 5, 6), 0, 0)
	time.Sleep(time.Second * 3)

	lcd.Clear()

	for i := 0; i < 70; i++ {
		lcd.DisplayString(fmt.Sprintf("%c", i%7), 0, 0)
		time.Sleep(time.Millisecond * 500)
	}
}
