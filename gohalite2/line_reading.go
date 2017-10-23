package gohalite2

import (
	"bufio"
	"os"
	"strconv"
)

// ---------------------------------------

type TokenParser struct {
	scanner		*bufio.Scanner
	count		int
}

func NewTokenParser() *TokenParser {
	ret := new(TokenParser)
	ret.scanner = bufio.NewScanner(os.Stdin)
	ret.scanner.Split(bufio.ScanWords)
	return ret
}

func (self *TokenParser) Int() int {
	self.scanner.Scan()
	ret, err := strconv.Atoi(self.scanner.Text())
	if err != nil {
		panic("TokenReader.NextInt(): Atoi failed at token " + strconv.Itoa(self.count))
	}
	self.count++
	return ret
}

func (self *TokenParser) Float() float64 {
	self.scanner.Scan()
	ret, err := strconv.ParseFloat(self.scanner.Text(), 64)
	if err != nil {
		panic("TokenReader.NextInt(): ParseFloat failed at token " + strconv.Itoa(self.count))
	}
	self.count++
	return ret
}
