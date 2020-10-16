package test

import (
	"github.com/netbrain/darknetw/cfg"
	"github.com/netbrain/darknetw/fs"
	"os"
	"path/filepath"
	"time"
)

func BootstrapTestEnvironment() (config *cfg.AppConfig, cleanup func()) {
	tmpdir := filepath.Join(os.TempDir(), time.Now().Format(cfg.TimeFormatFS))
	config = &cfg.AppConfig{
		NeuralNetworkConfig: &cfg.NeuralNetworkConfig{
			ConfigFile: filepath.Join(tmpdir, "network.cfg"),
			DataFile:   filepath.Join(tmpdir, "dataset.cfg"),
		},
		Storage:      tmpdir,
		DatasetSplit: 0.1,
	}
	cleanup = func() {
		_ = os.RemoveAll(config.Storage)
	}

	err := os.MkdirAll(config.Storage, 0755)
	if err != nil {
		panic(err)
	}

	//Copy files to temp dir
	for _, f := range []string{"dataset.cfg", "network.cfg", "names.txt", "train.txt", "valid.txt"} {
		err = fs.CopyFile(
			filepath.Join(fs.RuntimeDirectory(), "..", "example", f),
			filepath.Join(config.Storage, f),
		)
		if err != nil {
			panic(err)
		}
	}

	return
}
