package cfg

import (
	"github.com/gofrs/flock"
	"log"
	"path/filepath"
)

const TimeFormatFS = "02012006_150405"

type AppConfig struct {
	*ServerConfig
	*NeuralNetworkConfig
	Storage      string
	DatasetSplit float64 //0.0 - 1.0
}

type ServerConfig struct {
	Host string //http host
	Port string //http port
}

type NeuralNetworkConfig struct {
	ConfigFile  string //darknet config file
	WeightsFile string //darknet weights file
	DataFile    string //darknet data file
	Clear       bool   //will clear training statistics
}

func (c *AppConfig) TrainingLogPath() string {
	return filepath.Join(c.TrainingBasePath(), "train.log")
}

func (c *AppConfig) TrainingStatsPath() string {
	return filepath.Join(c.TrainingBasePath(), "train.json")
}

func (c *AppConfig) TrainingBasePath() string {
	return filepath.Join(c.Storage, "train")
}

func (c *AppConfig) ValidateStatsPath() string {
	return filepath.Join(c.Storage, "validate.json")
}

func (c *AppConfig) DatasetPath() string {
	return filepath.Join(c.Storage, "dataset")
}

func (c *AppConfig) LockTraining() *flock.Flock {
	return flock.New(filepath.Join(c.Storage, "train.lock"))
}

func (c *AppConfig) IsTraining() bool {
	lock := c.LockTraining()
	ok, err := lock.TryLock()
	if err != nil {
		log.Println(err)
		return false
	}
	defer lock.Unlock()
	return !ok
}
