package darknet

// #include <darknet.h>
import "C"
import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"unsafe"
)

type Image struct {
	data   []float32
	CImage C.image
}

func NewImageFromBytes(buf []byte) (*Image, error) {
	return NewImageFromReader(bytes.NewBuffer(buf))
}

func NewImageFromReader(reader io.Reader) (*Image, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return NewImage(img), nil
}

func NewImage(src image.Image) *Image {
	width := src.Bounds().Dx()
	height := src.Bounds().Dy()
	img := &Image{
		CImage: C.make_image(C.int(width), C.int(height), 3),
	}

	wh := width * height
	channels := 3
	length := wh * channels
	img.data = (*[1 << 30]float32)(unsafe.Pointer(img.CImage.data))[:length:length]

	for i := 0; i < wh; i++ {
		x := i % width
		y := i / width
		sc := length / 3
		r, g, b, _ := color.RGBAModel.Convert(src.At(x, y)).RGBA()
		img.data[i%sc] = float32(r) / 0xFFFF
		img.data[i%sc+sc] = float32(g) / 0xFFFF
		img.data[i%sc+sc*2] = float32(b) / 0xFFFF
	}

	return img
}

func (i *Image) Close() error {
	C.free_image(i.CImage)
	i.data = nil
	return nil
}
