package darknetcfg

import "path/filepath"

type DarknetInputFile string

func (d DarknetInputFile) String() string {
	return string(d)
}

func (d DarknetInputFile) StringTxt() string {
	return string(d)[0:len(d)-len(filepath.Ext(string(d)))] + ".txt"
}
