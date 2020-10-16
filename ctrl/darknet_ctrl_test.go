package ctrl

import (
	"bytes"
	"github.com/netbrain/darknetw/ctrl/multipart"
	"github.com/netbrain/darknetw/darknet"
	"github.com/netbrain/darknetw/test"
	"github.com/stretchr/testify/require"
	"image"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDarknetController_Label(t *testing.T) {
	config, cleanup := test.BootstrapTestEnvironment()
	defer cleanup()

	labels, err := darknet.ParseLabelFile(image.Rect(0, 0, 416, 416), "testdata/0.txt")
	require.NoError(t, err)

	ctrl := NewDarknetController(config)

	handler := CreateRouter(ctrl)
	body := &bytes.Buffer{}
	r := httptest.NewRequest("POST", "/api/v1/label", body)
	err = multipart.WriteMultipart(
		r,
		body,
		multipart.WithFormFile("image", "testdata/0.jpeg"),
		multipart.WithFormField("label", labels.JSON()),
	)
	require.NoError(t, err)

	response := Do(handler, r)
	require.Equal(t, http.StatusOK, response.StatusCode)
	//TODO test that label has been created in the storage directory
}
