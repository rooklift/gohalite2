package gohalite2

import (
	"encoding/json"
	"fmt"
	"os"
)

type Logfile struct {
	outfile         *os.File
	outfilename     string
	enabled         bool
	logged_once     map[string]bool
}

func NewLog(outfilename string, enabled bool) *Logfile {
	return &Logfile{
		nil,
		outfilename,
		enabled,
		make(map[string]bool),
	}
}

func (self *Logfile) Dump(format_string string, args ...interface{}) {

	if self == nil {
		return
	}

	if self.enabled == false {
		return
	}

	if self.outfile == nil {

		var err error

		if _, tmp_err := os.Stat(self.outfilename); tmp_err == nil {
			// File exists
			self.outfile, err = os.OpenFile(self.outfilename, os.O_APPEND|os.O_WRONLY, 0666)
		} else {
			// File needs creating
			self.outfile, err = os.Create(self.outfilename)
		}

		if err != nil {
			self.enabled = false
			return
		}
	}

	fmt.Fprintf(self.outfile, format_string, args...)
	fmt.Fprintf(self.outfile, "\r\n")                    // Because I use Windows...
}

func (self *Game) Log(format_string string, args ...interface{}) {
	self.logfile.Dump(format_string, args...)
}

func (self *Game) LogState() {
	b, _ := json.MarshalIndent(self, "", "  ")
	self.Log(string(b))
}
