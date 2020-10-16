package darknet

import (
	"github.com/stretchr/testify/require"
	"image"
	"testing"
)

func TestLabel_Yolo(t *testing.T) {
	yolo := "1 0.266827 0.545673 0.480769 0.288462"
	size := image.Rectangle{
		Max: image.Point{X: 416, Y: 416},
	}
	label, err := ParseLabel(size, yolo)
	require.NoError(t, err)
	result := label.Yolo(size)
	require.Equal(t, yolo, result)
}
