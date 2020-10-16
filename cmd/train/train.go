package train

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/netbrain/darknetw/cfg"
	"github.com/netbrain/darknetw/darknet"
	"github.com/netbrain/darknetw/darknet/darknetcfg"
	"github.com/netbrain/darknetw/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func Run(config *cfg.AppConfig) error {
	var err error
	lock := config.LockTraining()

	if ok, err := lock.TryLock(); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("training already started")
	}

	defer func() {
		e := lock.Unlock()
		if e != nil {
			err = e
			return
		}
		e = os.Remove(lock.Path())
		if e != nil {
			err = e
		}
	}()

	targetDir, datasetDir, err := createDirectoryLayout(config.Storage)
	if err != nil {
		return err
	}

	data, err := darknetcfg.ReadDataFile(config.DataFile)
	if err != nil {
		return err
	}

	for _, k := range []darknetcfg.DarknetDataKey{darknetcfg.Train, darknetcfg.Valid} {
		f := data.Get(k)
		if f == "" {
			continue
		}

		err = createAndLinkNewDatasetFromOriginal(data, k, f, targetDir, datasetDir, config.Storage)
		if err != nil {
			return err
		}
	}

	dataFileDst, configFileDst, weightsFileDst, err := modifyAndCopyDataFiles(config, data, targetDir, config.Storage)
	if err != nil {
		return err
	}

	err = os.Chdir(targetDir)
	if err != nil {
		return err
	}

	darknet.TrainDetectorCustom(
		dataFileDst,
		configFileDst,
		weightsFileDst,
		config.Clear,
		0, //TODO make this configurable
	)
	return nil
}

func modifyAndCopyDataFiles(config *cfg.AppConfig, dataFile *darknetcfg.DarknetData, targetDir, storageDir string) (dataFileDst, configFileDst, weightsFileDst string, err error) {
	if namesFile := dataFile.Get(darknetcfg.Names); namesFile != "" {
		dst := filepath.Join(targetDir, "names.txt")
		log.Printf("copying %s to %s", namesFile, dst)
		err = fs.CopyFile(namesFile, dst)
		if err != nil {
			return
		}

		relDst, err := filepath.Rel(targetDir, dst)
		if err != nil {
			log.Fatal(err)
		}

		dataFile.Set(darknetcfg.Names, relDst)
	}

	dataFile.Set(darknetcfg.Backup, "weights")

	dataFileDst, err = filepath.Abs(filepath.Join(targetDir, "dataset.cfg"))
	if err != nil {
		return
	}
	log.Printf("creating data file %s", dataFileDst)
	err = ioutil.WriteFile(dataFileDst, dataFile.Bytes(), 0644)
	if err != nil {
		return
	}

	configFileDst, err = filepath.Abs(filepath.Join(targetDir, "network.cfg"))
	if err != nil {
		return
	}
	log.Printf("copying config file to %s", configFileDst)
	err = fs.CopyFile(config.ConfigFile, configFileDst)
	if err != nil {
		return
	}

	if config.WeightsFile != "" {
		weightsFileDst, err = filepath.Abs(filepath.Join(targetDir, "starting.weights"))
		if err != nil {
			return
		}
		log.Printf("copying weights file to %s", weightsFileDst)
		err = fs.CopyFile(config.WeightsFile, weightsFileDst)
		if err != nil {
			return
		}
	}
	return
}

func createDirectoryLayout(basePath string) (targetDir string, datasetDir string, err error) {
	targetDir, err = filepath.Abs(filepath.Join(basePath, "train", time.Now().Format(cfg.TimeFormatFS)))
	if err != nil {
		return
	}
	log.Printf("creating a new training session @ %s", targetDir)

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return
	}

	datasetDir, err = filepath.Abs(filepath.Join(targetDir, "dataset"))
	if err != nil {
		return
	}
	err = os.MkdirAll(datasetDir, 0755)
	if err != nil {
		return
	}

	weightsDir, err := filepath.Abs(filepath.Join(targetDir, "weights"))
	if err != nil {
		return
	}
	err = os.MkdirAll(weightsDir, 0755)
	if err != nil {
		return
	}
	return
}

func createAndLinkNewDatasetFromOriginal(data *darknetcfg.DarknetData, k darknetcfg.DarknetDataKey, file, targetDir, datasetDir, storageDir string) error {
	datasetFile := filepath.Join(targetDir, fmt.Sprintf("%s.txt", k))
	relDatasetFile, err := filepath.Rel(targetDir, datasetFile)
	if err != nil {
		return err
	}
	data.Set(k, relDatasetFile)
	log.Printf("creating dataset file %s", datasetFile)
	fh, err := os.Create(datasetFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(buf))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		df := darknetcfg.DarknetInputFile(scanner.Text())
		if !filepath.IsAbs(df.String()) {
			df = darknetcfg.DarknetInputFile(filepath.Join(storageDir, string(df)))
		}
		for i, f := range []string{df.String(), df.StringTxt()} {
			dst := filepath.Join(datasetDir, filepath.Base(f))
			log.Printf("hardlinking %s to %s", f, dst)

			if err := os.Link(f, dst); err != nil {
				return err
			}
			if i%2 == 0 {
				relDst, err := filepath.Rel(targetDir, dst)
				if err != nil {
					return err
				}
				_, err = fh.WriteString(fmt.Sprintln(relDst))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
