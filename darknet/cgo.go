package darknet

// #cgo CFLAGS: -I${SRCDIR}/../include -I/usr/local/cuda/include
// #cgo LDFLAGS: -L${SRCDIR}/../lib -Wl,-rpath,$ORIGIN/lib -Wl,-rpath,${SRCDIR}/../lib -ldarknet
import "C"
