package ctrl

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	. "github.com/netbrain/darknetw/api"
	"github.com/netbrain/darknetw/cfg"
	"github.com/netbrain/darknetw/darknet"
	"github.com/netbrain/darknetw/darknet/darknetcfg"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type DarknetController struct {
	Network *darknet.Network
	*cfg.AppConfig
	validationPool sync.Pool //locking mechanism
}

func (c *DarknetController) Routes() Routes {
	return Routes{
		"/api/v1/predict": {
			POST: HandlerFn(c.Predict),
		},
		"/api/v1/label": {
			POST: HandlerFn(c.Label),
		},
		"/api/v1/train": {
			POST: HandlerFn(c.StartTraining),
			GET:  HandlerFn(c.ReportTrainingStatistics),
		},
		"/api/v1/accuracy": {
			GET:    HandlerFn(c.ReportAccuracyStatistics),
			DELETE: HandlerFn(c.ClearAccuracyStatistics),
		},
	}
}

func NewDarknetController(config *cfg.AppConfig) *DarknetController {
	controller := &DarknetController{
		AppConfig: config,
	}
	controller.validationPool.Put(struct{}{})
	return controller
}

func (c *DarknetController) Predict(ctx Context) Response {
	if c.Network == nil {
		c.Network = darknet.LoadNetwork(c.ConfigFile, c.DataFile, c.WeightsFile)
	}

	reader, err := ReadMultipart(ctx.Request)
	if err != nil {
		return BadRequest()
	}
	var response []PredictResponse
	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
			return Error(err)
		}

		response = append(response, PredictResponse{
			File:        part.FileName(),
			Name:        part.FormName(),
			ContentType: part.Header.Get("Content-Type"),
		})

		img, err := darknet.NewImageFromReader(part)
		if err != nil {
			return Error(err)
		}

		detections := c.Network.DetectImage(img)
		_ = img.Close()

		var rDetections []Detection
		for _, detection := range detections {
			rDetections = append(rDetections, Detection{
				X1:         int(detection.X1),
				Y1:         int(detection.Y1),
				X2:         int(detection.X2),
				Y2:         int(detection.Y2),
				Class:      detection.Class,
				ClassName:  detection.ClassName,
				Confidence: detection.Confidence,
			})
		}
		response[len(response)-1].Detections = rDetections
	}

	return JSON(response)
}

func (c *DarknetController) getDatasetToUse() (func() string, error) {
	dataFile, err := darknetcfg.ReadDataFile(c.DataFile)
	if err != nil {
		return nil, err
	}
	trainFile := dataFile.Get(darknetcfg.Train)
	validFile := dataFile.Get(darknetcfg.Valid)

	trainCount := countLines(trainFile)
	validCount := countLines(validFile)
	if validCount < 0 || trainCount < 0 {
		return nil, fmt.Errorf("failed to count number of lines in train/valid")
	}

	return func() string {
		if float64(trainCount)/float64(validCount)/100 < c.DatasetSplit {
			validCount++
			return validFile
		}
		trainCount++
		return trainFile
	}, nil
}

func (c *DarknetController) Label(ctx Context) Response {
	reader, err := ReadMultipart(ctx.Request)
	if err != nil {
		return BadRequest()
	}

	err = os.MkdirAll(c.DatasetPath(), 0755)
	if err != nil {
		return Error(err)
	}

	datasetToUse, err := c.getDatasetToUse()
	if err != nil {
		return Error(err)
	}

	var imgBuf []byte
	var img image.Image
	var size image.Rectangle
	var ext string
	for i := 0; ; i++ {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
			return Error(err)
		}
		if i%2 == 0 {
			//image
			imgBuf, err = ioutil.ReadAll(part)
			if err != nil {
				return Error(err)
			}
			img, ext, err = image.Decode(bytes.NewBuffer(imgBuf))
			if err != nil {
				return Error(err)
			}
			size = img.Bounds()
		} else {
			//json
			var labels []*Label
			err := json.NewDecoder(part).Decode(&labels)
			if err != nil {
				return Error(err)
			}

			imgMd5 := fmt.Sprintf("%x", md5.Sum(imgBuf))
			imgDst := filepath.Join(c.DatasetPath(), imgMd5+"."+ext)
			txtDst := filepath.Join(c.DatasetPath(), imgMd5+".txt")
			err = ioutil.WriteFile(imgDst, imgBuf, 0644)
			if err != nil {
				return Error(err)
			}

			dataset := datasetToUse()
			outFile, err := os.OpenFile(dataset, os.O_RDWR, 0644)
			if err != nil {
				return Error(err)
			}
			fInfo, err := outFile.Stat()
			if err != nil {
				return Error(err)
			}
			if fInfo.Size() > 0 {
				_, err = outFile.Seek(-1, io.SeekEnd)
				if err != nil {
					return Error(err)
				}
				nlbuf := []byte{0}
				_, err = outFile.Read(nlbuf)
				if err != nil {
					return Error(err)
				}
				if string(nlbuf) != "\n" {
					_, err = outFile.WriteString("\n")
					if err != nil {
						return Error(err)
					}
				}
			}
			relImgDst, err := filepath.Rel(filepath.Dir(c.DatasetPath()), imgDst)
			if err != nil {
				return Error(err)
			}
			_, err = outFile.WriteString(relImgDst)
			if err != nil {
				return Error(err)
			}
			_ = outFile.Close()

			var yoloLabels []string
			for _, label := range labels {
				yoloLabels = append(yoloLabels, label.Yolo(size))
			}

			err = ioutil.WriteFile(txtDst, []byte(strings.Join(yoloLabels, "\n")), 0644)
			if err != nil {
				return Error(err)
			}
		}
	}
	return OK()
}

func (c *DarknetController) StartTraining(ctx Context) Response {
	if c.IsTraining() {
		return Status(http.StatusServiceUnavailable)
	}

	data := &TrainingRequest{
		Data:    c.DataFile,
		Config:  c.ConfigFile,
		Weights: c.WeightsFile,
		Clear:   c.Clear,
	}
	if ctx.Request.ContentLength > 0 {
		err := json.NewDecoder(ctx.Request.Body).Decode(data)
		if err != nil {
			return BadRequest()
		}
	}

	args := []string{"train", "--data", data.Data, "--config", data.Config, "--weights", data.Weights}
	if data.Clear {
		args = append(args, "--clear")
	}
	cmd := exec.Command(os.Args[0], args...)
	r, w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Start()
	if err != nil {
		return Error(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go c.readTrainingOutput(r, &wg)
	wg.Wait() //wait until first iteration

	return Redirect(ctx.Request.RequestURI, http.StatusSeeOther)
}

func (c *DarknetController) ReportTrainingStatistics(_ Context) Response {
	if !c.IsTraining() {
		return ErrorString(
			http.StatusNotFound,
			"currently not running a training session",
		)
	}

	var err error
	var exists bool
	var data []byte
	for i := 0; i < 5; i++ {
		data, err = ioutil.ReadFile(c.TrainingStatsPath())
		if os.IsNotExist(err) {
			time.Sleep(time.Second)
			continue
		}
		exists = true
		break
	}

	if !exists {
		return ErrorString(
			http.StatusServiceUnavailable,
			"training has not been started yet, retry in 5 seconds.",
			WithHeader("Retry-After", "5"),
		)
	}

	return JSONRaw(data)
}

func (c *DarknetController) ClearAccuracyStatistics(_ Context) Response {
	lock := c.validationPool.Get()
	if lock == nil {
		return Status(http.StatusServiceUnavailable)
	}
	defer c.validationPool.Put(lock)

	matches, err := filepath.Glob(filepath.Join(c.TrainingBasePath(), "**/weights/*.weights"))
	if err != nil {
		return Error(err)
	}

	accuracyStats := map[string]Accuracy{}
	for _, weightFile := range matches {
		baseDir := filepath.Dir(filepath.Dir(weightFile))
		args := []string{
			"validate",
			"--data", filepath.Join(baseDir, "dataset.cfg"),
			"--config", filepath.Join(baseDir, "network.cfg"),
			"--weights", weightFile,
		}
		cmd := exec.Command(os.Args[0], args...)
		cmd.Dir = baseDir
		buf, err := cmd.CombinedOutput()
		if err != nil {
			return Error(err)
		}

		relPath, err := filepath.Rel(c.Storage, weightFile)
		if err != nil {
			return Error(err)
		}
		accuracyStats[relPath] = c.readValidationOutput(bytes.NewBuffer(buf))
	}

	fh, err := os.Create(c.ValidateStatsPath())
	if err != nil {
		return Error(err)
	}
	defer fh.Close()
	if err := json.NewEncoder(fh).Encode(accuracyStats); err != nil {
		return Error(err)
	}
	return Redirect("/api/v1/accuracy", http.StatusSeeOther)
}

func (c *DarknetController) ReportAccuracyStatistics(_ Context) Response {
	buf, err := ioutil.ReadFile(c.ValidateStatsPath())
	if err != nil {
		return Error(err)
	}

	return JSONRaw(buf)
}

func countLines(filename string) int {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return -1
	}

	return bytes.Count(buf, []byte("\n"))
}

func (c *DarknetController) readTrainingOutput(r io.Reader, wg *sync.WaitGroup) {
	var once sync.Once
	defer once.Do(wg.Done)

	flh, err := os.Create(c.TrainingLogPath())
	if err != nil {
		log.Println(err)
		return
	}
	defer flh.Close()
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	statsRe := regexp.MustCompile(`(\d+): ([0-9.]+), ([0-9.]+) avg loss, ([0-9.]+) rate, ([0-9.]+) seconds, (\d+) images, ([+-]?[0-9.]+) hours left`)
	mapRe := regexp.MustCompile(`Last accuracy mAP@([0-9\.]+) = ([0-9\.]+) %, best = ([0-9\.]+) %`)
	var stats struct {
		Iteration       int     `json:"iteration"`
		Loss            float64 `json:"loss"`
		AvgLoss         float64 `json:"avgLoss"`
		CurrentRate     float64 `json:"currentRate"`
		ElapsedSeconds  float64 `json:"elapsedSeconds"`
		Images          int     `json:"images"`
		HoursLeft       float64 `json:"hoursLeft"`
		MapIOUThreshold float64 `json:"mapIouThreshold"`
		MapLast         float64 `json:"mapLast"`
		MapBest         float64 `json:"mapBest"`
		Target          string  `json:"target"`
	}

	var directoryCreated bool
	for scanner.Scan() {
		func() {
			line := strings.TrimSpace(scanner.Text())
			if !directoryCreated && strings.HasPrefix(line, "creating a new training session") {
				files, err := ioutil.ReadDir(c.TrainingBasePath())
				if err != nil {
					log.Println(err)
					return
				}
				var path string
				var max time.Time
				for _, f := range files {
					if max.After(f.ModTime()) {
						continue
					}
					max = f.ModTime()
					path = filepath.Join("train", f.Name())
				}
				stats.Target = path
			}
			if strings.HasPrefix(line, "Last accuracy mAP@") {
				matches := mapRe.FindAllStringSubmatch(line, -1)
				stats.MapIOUThreshold = tryToFloat(matches[0][1])
				stats.MapLast = tryToFloat(matches[0][2]) / 100
				stats.MapBest = tryToFloat(matches[0][3]) / 100
			}
			if strings.HasSuffix(line, "hours left") {
				fjh, err := os.Create(c.TrainingStatsPath())
				if err != nil {
					log.Println(err)
					return
				}
				defer fjh.Close()
				defer once.Do(wg.Done)
				matches := statsRe.FindAllStringSubmatch(line, -1)
				encoder := json.NewEncoder(fjh)

				stats.Iteration = tryToInt(matches[0][1])
				stats.Loss = tryToFloat(matches[0][2])
				stats.AvgLoss = tryToFloat(matches[0][3])
				stats.CurrentRate = tryToFloat(matches[0][4])
				stats.ElapsedSeconds = tryToFloat(matches[0][5])
				stats.Images = tryToInt(matches[0][6])
				stats.HoursLeft = tryToFloat(matches[0][7])

				err = encoder.Encode(stats)
				if err != nil {
					log.Println(err)
					return
				}
			}
			_, err := flh.WriteString(fmt.Sprintln(line))
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
}

func (c *DarknetController) readValidationOutput(r io.Reader) Accuracy {
	scanner := bufio.NewScanner(r)

	//darknet returns some output with \r terminated lines, this is why we can't use bufio.ScanLines
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\r'); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil

	})

	classesRe := regexp.MustCompile(`class_id = (.*?), name = (.*?), ap = (.*?)%.*\(TP = (.*?), FP = (.*?)\)`)
	metricsRe := regexp.MustCompile(`for conf_thresh = (.*?), precision = (.*?), recall = (.*?), F1-score = (.*?)$`)
	countsRe := regexp.MustCompile(`for conf_thresh = .*?, TP = (.*?), FP = (.*?), FN = (.*?), average IoU = (.*?) %`)
	mapRe := regexp.MustCompile(`mean average precision \(mAP@(.*?)\) = (.*?),.*$`)

	var stats Accuracy
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "class_id =") {
			matches := classesRe.FindAllStringSubmatch(line, -1)
			stats.Classes = append(stats.Classes, ClassAccuracy{
				ID:               tryToInt(matches[0][1]),
				Name:             matches[0][2],
				AveragePrecision: tryToFloat(matches[0][3]) / 100,
				TruePositives:    tryToInt(matches[0][4]),
				FalsePositives:   tryToInt(matches[0][5]),
			})
		} else if strings.HasPrefix(line, "for conf_thresh =") {
			if metricsRe.MatchString(line) {
				matches := metricsRe.FindAllStringSubmatch(line, -1)
				stats.Threshold = tryToFloat(matches[0][1])
				stats.Precision = tryToFloat(matches[0][2])
				stats.Recall = tryToFloat(matches[0][3])
				stats.F1 = tryToFloat(matches[0][4])
			} else if countsRe.MatchString(line) {
				matches := countsRe.FindAllStringSubmatch(line, -1)
				stats.TruePositives = tryToInt(matches[0][1])
				stats.FalsePositives = tryToInt(matches[0][2])
				stats.FalseNegatives = tryToInt(matches[0][3])
				stats.AverageIoU = tryToFloat(matches[0][4]) / 100
			} else {
				log.Println("warning: no regex match found")
			}
		} else if strings.HasPrefix(line, "mean average precision") {
			matches := mapRe.FindAllStringSubmatch(line, -1)
			stats.MapIOUThreshold = tryToFloat(matches[0][1])
			stats.Map = tryToFloat(matches[0][2])
		}
	}

	return stats
}

type TrainingRequest struct {
	Data    string `json:"data"`
	Config  string `json:"config"`
	Weights string `json:"weights"`
	Clear   bool   `json:"clear"`
}

type PredictResponse struct {
	File        string      `json:"file"`
	Name        string      `json:"name"`
	ContentType string      `json:"contentType"`
	Detections  []Detection `json:"detections"`
}

type Detection struct {
	Class      int     `json:"class"`
	ClassName  string  `json:"className"`
	Confidence float32 `json:"confidence"`
	X1         int     `json:"x1"`
	Y1         int     `json:"y1"`
	X2         int     `json:"x2"`
	Y2         int     `json:"y2"`
}

type Bbox struct {
}

type Accuracy struct {
	Classes         []interface{} `json:"classes"`
	Threshold       float64       `json:"threshold"`
	Precision       float64       `json:"precision"`
	Recall          float64       `json:"recall"`
	F1              float64       `json:"f1"`
	TruePositives   int           `json:"tp"`
	FalsePositives  int           `json:"fp"`
	FalseNegatives  int           `json:"fn"`
	AverageIoU      float64       `json:"averageIoU"`
	Map             float64       `json:"map"`
	MapIOUThreshold float64       `json:"mapIouThreshold"`
}

type ClassAccuracy struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	AveragePrecision float64 `json:"ap"`
	TruePositives    int     `json:"tp"`
	FalsePositives   int     `json:"fp"`
}

type Label struct {
	X1    float64 `json:"x1"`
	Y1    float64 `json:"y1"`
	X2    float64 `json:"x2"`
	Y2    float64 `json:"y2"`
	Class int     `json:"class"`
}

//TODO handle multiple classes? [1, x_center, y_center, width, height, 1, 0, 1, 0, 0]
func (l *Label) Yolo(size image.Rectangle) string {
	w := (l.X2 - l.X1) / float64(size.Max.X)
	h := (l.Y2 - l.Y1) / float64(size.Max.Y)
	return fmt.Sprintf("%d %f %f %f %f",
		l.Class,
		l.X1/float64(size.Max.X)+w/2,
		l.Y1/float64(size.Max.Y)+h/2,
		w,
		h,
	)
}
