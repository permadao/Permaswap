package logger

import "github.com/inconshreveable/log15"

func New(module string) (log log15.Logger) {
	log = log15.New("module", module)
	log.SetHandler(log15.LvlFilterHandler(log15.LvlInfo, log15.StdoutHandler))
	return
}
