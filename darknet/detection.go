package darknet

type Detection struct {
	Label
	ClassName  string
	Confidence float32
}
