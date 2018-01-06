# Golang HD44780 LCD Library

## Compatibility
Tested with both 2 * 16 and 4 * 20 screens on a Raspberry Pi Zero.

## Motivation
I couldn't get any golang hd44780 libraries to output a degree symbol ˚ correctly (char 223), yet it worked fine with a 
[python library](http://www.recantha.co.uk/blog/?p=4849). I couldn't work out where the bug was so decided to implement 
the python lib in go, it worked (I still feel like the problem was my side somehow). I then kept going, refactoring and 
implementing the best features and patterns I found in other implementations. Much of what I ended up with came from 
[kidoman/embd](https://github.com/kidoman/embd/blob/master/controller/hd44780/hd44780.go).

## Reading data from HD44780
I spent quite a long time trying to read data via I²C, didn't manage to get it working but have left in some code I 
tried to use.

## Docs

https://godoc.org/github.com/j0hnsmith/hd44780

## Example usage

https://godoc.org/github.com/j0hnsmith/hd44780/#example_

