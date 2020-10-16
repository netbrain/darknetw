package darknet

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"strconv"
	"strings"
)

type Labels []*Label

func (l Labels) Yolo(size image.Rectangle) string {
	var yolos []string
	for _, ll := range l {
		yolos = append(yolos, ll.Yolo(size))
	}
	return strings.Join(yolos, "\n")
}

func (l Labels) JSON() []byte {
	buf, err := json.Marshal(l)
	if err != nil {
		return nil
	}
	return buf
}

type Label struct {
	X1    float64 `json:"x1"`
	X2    float64 `json:"x2"`
	Y1    float64 `json:"y1"`
	Y2    float64 `json:"y2"`
	Class int     `json:"class"`
}

func ParseLabelFile(size image.Rectangle, file string) (Labels, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.Trim(string(buf), "\n"), "\n")
	var labels []*Label
	for _, l := range lines {
		label, err := ParseLabel(size, l)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	return labels, nil
}

func ParseLabel(size image.Rectangle, yolo string) (*Label, error) {
	yolo = strings.TrimSpace(yolo)
	parts := strings.Split(yolo, " ")
	if len(parts) != 5 {
		return nil, fmt.Errorf("incorrect format")
	}

	class, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	x, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}

	y, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, err
	}

	w, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return nil, err
	}

	h, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return nil, err
	}

	sw := float64(size.Max.X)
	sh := float64(size.Max.Y)
	return &Label{
		X1:    x*sw - w*sw/2,
		Y1:    y*sh - h*sh/2,
		X2:    x*sw + w*sw/2,
		Y2:    y*sh + h*sh/2,
		Class: class,
	}, nil
}

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

func (l *Label) Rectangle() image.Rectangle {
	return image.Rect(int(l.X1), int(l.Y1), int(l.X2), int(l.Y2))
}

func (l *Label) JSON() []byte {
	buf, err := json.Marshal(l)
	if err != nil {
		return nil
	}
	return buf
}
