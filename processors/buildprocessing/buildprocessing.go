package buildprocessing

import (
	"lmcli/processors"
)


type processing struct {

}

func (b *processing) ProcessCmd(cmd []string) {

}

func New() processors.CmdProcessor {
	return new(processing)
}