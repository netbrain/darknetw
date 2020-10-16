package validate

import (
	"github.com/netbrain/darknetw/cfg"
	"github.com/netbrain/darknetw/darknet"
)

func Run(config *cfg.AppConfig) (err error) {
	darknet.ValidateDetectorMap(
		config.DataFile,
		config.ConfigFile,
		config.WeightsFile,
	)
	return
}
