package darknetcfg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"path/filepath"
)

type DarknetDataKey string

const (
	Classes DarknetDataKey = "classes"
	Train   DarknetDataKey = "train"
	Valid   DarknetDataKey = "valid"
	Names   DarknetDataKey = "names"
	Backup  DarknetDataKey = "backup"
)

// ReadDataFile parses a darknet data file and additionally stores the location of this file. It uses this information
// to resolve files with an absolute path if it isn't already.
func ReadDataFile(fp string) (*DarknetData, error) {
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	dnd, err := ReadData(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	dnd.path, err = filepath.Abs(fp)
	if err != nil {
		return nil, err
	}
	return dnd, nil
}

// ReadData parses darknet data, doesn't do any additional filepath resolving as ReadDataFile does
func ReadData(r io.Reader) (*DarknetData, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(scanData)

	out := &DarknetData{}
	var key DarknetDataKey
	for i := 0; ; i++ {
		if !scanner.Scan() {
			break
		}

		if i%2 != 0 {
			out.Set(key, scanner.Text())
			continue
		}

		key = DarknetDataKey(scanner.Text())
	}

	return out, nil
}

type DarknetData struct {
	m     map[DarknetDataKey]string
	order []DarknetDataKey
	path  string
}

// Get returns the value for the given key, if the key is determined to contain a file path and the datafile path is
// available, then it will resolve these values as absolute file paths.
func (d *DarknetData) Get(key DarknetDataKey) string {
	if d.m == nil {
		return ""
	}
	switch key {
	case Names:
		fallthrough
	case Valid:
		fallthrough
	case Train:
		fallthrough
	case Backup:
		if !filepath.IsAbs(d.m[key]) {
			return filepath.Join(filepath.Dir(d.path), d.m[key])
		}
	}
	return d.m[key]
}

// Set will set the value by the given key
func (d *DarknetData) Set(key DarknetDataKey, value string) {
	if d.m == nil {
		d.m = map[DarknetDataKey]string{}
	}
	d.m[key] = value
	var exists bool
	for _, o := range d.order {
		if o == key {
			exists = true
			break
		}
	}
	if exists {
		return
	}
	d.order = append(d.order, key)
}

func (d *DarknetData) Bytes() (buf []byte) {
	for i := 0; i < len(d.order); i++ {
		buf = append(buf, []byte(fmt.Sprintf("%s = %s\n", d.order[i], d.m[d.order[i]]))...)
	}
	buf = bytes.TrimRight(buf, "\n")
	return
}

func (d *DarknetData) String() string {
	return string(d.Bytes())
}

func scanData(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	eq := bytes.IndexByte(data, '=')
	nl := bytes.IndexByte(data, '\n')
	i := int(math.Min(float64(eq), float64(nl)))

	if i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, bytes.TrimSpace(data[0:i]), nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), bytes.TrimSpace(data), nil
	}
	// Request more data.
	return 0, nil, nil
}
