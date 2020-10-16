package darknettest

import (
	"bytes"
	"fmt"
	"github.com/netbrain/darknetw/darknet"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

// GenerateTestDataset creates a dataset with rectangles and circles and writes to the specified output directory in
// yolo format
func GenerateTestDataset(outputDir string, images int, seed int) error {
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	random := rand.New(rand.NewSource(int64(seed)))

	img := image.NewRGBA(image.Rectangle{
		Max: image.Point{
			X: 416,
			Y: 416,
		},
	})

	for i := 0; i < images; i++ {
		if random.Float64() > 0.5 {
			random.Read(img.Pix)
		} else {
			img.Pix = bytes.Repeat([]byte{
				byte(random.Intn(255)),
				byte(random.Intn(255)),
				byte(random.Intn(255)),
				byte(random.Intn(255)),
			}, 4*img.Bounds().Max.X*img.Bounds().Max.Y)
		}

		var labels []*darknet.Label
		for i := 0; i < random.Intn(10); i++ {
			if random.Float64() > 0.5 {
				center := image.Pt(random.Intn(416), random.Intn(416))
				radius := random.Intn(75) + 10
				labels = append(labels, &darknet.Label{
					X1:    float64(center.X - radius - 1),
					Y1:    float64(center.Y - radius - 1),
					X2:    float64(center.X + radius + 1),
					Y2:    float64(center.Y + radius + 1),
					Class: 0,
				})
				CircleFill(img, center, radius, randomColor(random))
			} else {
				x1 := random.Float64() * 416
				y1 := random.Float64() * 416
				x2 := random.Float64() * 416
				y2 := random.Float64() * 416
				x1, x2 = math.Min(x1, x2), math.Max(x1, x2)
				y1, y2 = math.Min(y1, y2), math.Max(y1, y2)

				if x2-x1 < 10 {
					x2 += 10
				}

				if y2-y1 < 10 {
					y2 += 10
				}

				if x2-x1 > 200 {
					x2 = x1 + 200
				}

				if y2-y1 > 200 {
					y2 = y1 + 200
				}
				labels = append(labels, &darknet.Label{
					X1:    x1,
					Y1:    y1,
					X2:    x2,
					Y2:    y2,
					Class: 1,
				})
				RectFill(img, labels[len(labels)-1].Rectangle(), randomColor(random))
			}
		}
		f, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%d.jpeg", i)))
		if err != nil {
			return err
		}
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			return err
		}
		f.Close()

		var buf []byte
		for _, l := range labels {
			buf = append(buf, []byte(l.Yolo(img.Bounds())+"\n")...)
		}
		err = ioutil.WriteFile(filepath.Join(outputDir, fmt.Sprintf("%d.txt", i)), buf, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func randomColor(random *rand.Rand) color.RGBA {
	return color.RGBA{
		R: uint8(random.Intn(255)),
		G: uint8(random.Intn(255)),
		B: uint8(random.Intn(255)),
		A: uint8(random.Intn(255)),
	}
}
