package darknet

import "unsafe"

// #include <darknet.h>
import "C"

func LoadNetwork(configFile, dataFile, weightsFile string) *Network {
	return LoadNetworkCustom(configFile, dataFile, weightsFile, 1)
}

func LoadNetworkCustom(configFile, dataFile, weightsFile string, batchSize int) *Network {
	net := &Network{
		CNetwork:  C.load_network_custom(C.CString(configFile), C.CString(weightsFile), C.int(0), C.int(batchSize)),
		CMetadata: C.get_metadata(C.CString(dataFile)),
	}
	classes := int(net.CMetadata.classes)
	cnames := (*[1 << 30]*C.char)(unsafe.Pointer(net.CMetadata.names))[:classes:classes]
	names := make([]string, classes, classes)
	for i, cname := range cnames {
		names[i] = C.GoString(cname)
	}
	net.ClassNames = names
	return net
}

func TrainDetector(dataCfg, cfgFile, weightFile string) {
	TrainDetectorCustom(dataCfg, cfgFile, weightFile, false)
}

func TrainDetectorCustom(dataCfg, cfgFile, weightFile string, clear bool, gpus ...int) {
	defer C.fflush(C.stdout)
	if gpus == nil {
		gpus = []int{0}
	}
	var cWeightFile *C.char
	if weightFile != "" {
		cWeightFile = C.CString(weightFile)
	}
	//LIB_API void train_detector(char *datacfg, char *cfgfile, char *weightfile, int *gpus, int ngpus, int clear, int dont_show, int calc_map, int mjpeg_port, int show_imgs, int benchmark_layers, char* chart_path);
	C.train_detector(C.CString(dataCfg), C.CString(cfgFile), cWeightFile, (*C.int)(unsafe.Pointer(&gpus[0])), C.int(len(gpus)), boolToCInt(clear), C.int(1), C.int(1), C.int(0), C.int(0), C.int(0), nil)
}

func ValidateDetectorMap(dataCfg, cfgFile, weightFile string) {
	ValidateDetectorCustom(dataCfg, cfgFile, weightFile, 0.25, 0.5, 0, false, nil)
}

func ValidateDetectorCustom(dataCfg, cfgFile, weightFile string, threshCalcAvgIou, iouThresh float32, mapPoints int, letterBox bool, network *Network) {
	defer C.fflush(C.stdout)
	var existingNetwork *C.network
	if network != nil {
		existingNetwork = network.CNetwork
	}
	//LIB_API float validate_detector_map(char *datacfg, char *cfgfile, char *weightfile, float thresh_calc_avg_iou, const float iou_thresh, const int map_points, int letter_box, network *existing_net);
	C.validate_detector_map(C.CString(dataCfg), C.CString(cfgFile), C.CString(weightFile), C.float(threshCalcAvgIou), C.float(iouThresh), C.int(mapPoints), boolToCInt(letterBox), existingNetwork)
}

func boolToCInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}
