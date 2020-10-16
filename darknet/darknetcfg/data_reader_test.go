package darknetcfg

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestReadData(t *testing.T) {
	data, err := ReadData(bytes.NewBuffer([]byte(`
		classes=6
		train= example/train.txt
		valid =example/test.txt
		names = example/obj.names
	`)))
	require.NoError(t, err)

	require.Equal(t, "6", data.Get(Classes))
	require.Equal(t, "example/train.txt", data.Get(Train))
	require.Equal(t, "example/test.txt", data.Get(Valid))
	require.Equal(t, "example/obj.names", data.Get(Names))
	require.Equal(t, data.String(),
		"classes = 6\n"+
			"train = example/train.txt\n"+
			"valid = example/test.txt\n"+
			"names = example/obj.names")
}

func TestReadDataFile(t *testing.T) {
	dataFile, err := ioutil.TempFile("", "*")
	require.NoError(t, err)

	_, err = dataFile.Write([]byte(`
		classes = 6
		train = train.txt
		valid = valid.txt
		names = names.txt
	`))
	require.NoError(t, err)
	require.NoError(t, dataFile.Close())

	data, err := ReadDataFile(dataFile.Name())
	require.NoError(t, err)

	require.Equal(t, "6", data.Get(Classes))
	require.Equal(t, filepath.Join(filepath.Dir(dataFile.Name()), "train.txt"), data.Get(Train))
	require.Equal(t, filepath.Join(filepath.Dir(dataFile.Name()), "valid.txt"), data.Get(Valid))
	require.Equal(t, filepath.Join(filepath.Dir(dataFile.Name()), "names.txt"), data.Get(Names))
	require.Equal(t, data.String(),
		"classes = 6\n"+
			"train = train.txt\n"+
			"valid = valid.txt\n"+
			"names = names.txt")
}

func TestDarknetData_Set(t *testing.T) {
	data := &DarknetData{}
	data.Set(Classes, "6")
	data.Set(Train, "example/train.txt")
	data.Set(Valid, "example/valid.txt")
	data.Set(Names, "example/obj.names")
	data.Set(Backup, "backup/")
	require.Equal(t,
		"classes = 6\n"+
			"train = example/train.txt\n"+
			"valid = example/valid.txt\n"+
			"names = example/obj.names\n"+
			"backup = backup/",
		data.String(),
	)
}
