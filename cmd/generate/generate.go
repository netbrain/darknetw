package generate

import (
	"github.com/netbrain/darknetw/darknet/darknettest"
)

func Run(outputDir string, images, seed int) error {
	return darknettest.GenerateTestDataset(outputDir, images, seed)
}
