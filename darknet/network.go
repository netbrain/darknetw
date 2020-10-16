package darknet

// #include <darknet.h>
import "C"
import (
	"sync"
	"unsafe"
)

type Network struct {
	mu         sync.Mutex
	CNetwork   *C.network
	CMetadata  C.metadata
	ClassNames []string
}

func (n *Network) DetectImage(image *Image) []*Detection {
	return n.DetectImageCustom(image, 0.5, 0.5, 0.45)
}

func (n *Network) DetectImageCustom(img *Image, thresh, hierThresh, nms float32) []*Detection {
	n.mu.Lock()
	defer n.mu.Unlock()
	width := C.network_width(n.CNetwork)
	height := C.network_height(n.CNetwork)
	num := C.int(0)

	resizedImg := C.resize_image(img.CImage, width, height)
	defer C.free_image(resizedImg)

	C.network_predict(*n.CNetwork, resizedImg.data)

	detections := C.get_network_boxes(n.CNetwork, width, height, C.float(thresh), C.float(hierThresh), nil, 1, (*C.int)(&num), 0)
	defer C.free_detections(detections, num)

	if nms > 0 {
		C.do_nms_sort(detections, num, n.CMetadata.classes, C.float(nms))
	}

	cdets := (*[1 << 30]C.detection)(unsafe.Pointer(detections))[:num:num]
	var dets []*Detection
	for _, cdet := range cdets {
		classes := int(cdet.classes)
		probs := (*[1 << 30]C.float)(unsafe.Pointer(cdet.prob))[:classes:classes]

		for i := range probs {
			prob := float32(probs[i])
			if prob < thresh {
				continue
			}
			dets = append(dets, &Detection{
				Label: Label{
					X1:    float64((float32(cdet.bbox.x) - float32(cdet.bbox.w)/2.0) * float32(img.CImage.w)),
					Y1:    float64((float32(cdet.bbox.y) - float32(cdet.bbox.h)/2.0) * float32(img.CImage.h)),
					X2:    float64((float32(cdet.bbox.x) + float32(cdet.bbox.w)/2.0) * float32(img.CImage.w)),
					Y2:    float64((float32(cdet.bbox.y) + float32(cdet.bbox.h)/2.0) * float32(img.CImage.h)),
					Class: i,
				},
				ClassName:  n.ClassNames[i],
				Confidence: prob,
			})
		}
	}

	return dets
}

func (n *Network) Close() error {
	C.free_network(*n.CNetwork)
	return nil
}
