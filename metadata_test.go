package churl

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/object88/churl/internal/request"
	"github.com/object88/churl/mocks"
)

func Test_Metadata_Do(t *testing.T) {
	knownGoodResponse := `[{
		"name": "foo",
		"version": "1.0.0",
		"apiVersion": "v2",
		"appVersion": "1.0.1",
		"urls": [
			"charts/foo-1.0.0.tgz"
		],
		"created": "2019-11-21T05:44:14.924904011Z",
		"digest": "2c1e7190eadba25280cd08bacb40ccb9afb78d029d8ed4f371d8b490e5303c6e"
	}]`

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpResp := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(knownGoodResponse)),
		Status:     "OK",
		StatusCode: http.StatusOK,
	}

	mrt := mocks.NewMockRoundTripper(ctrl)
	mrt.EXPECT().RoundTrip(gomock.Any()).Return(httpResp, nil).Times(1)

	req, _ := request.NewRequest("localhost:1234")
	req.Transport = mrt
	mr := &MetadataReader{
		req: req,
	}

	_, err := mr.Do("foo")
	if err != nil {
		t.Errorf("Failed to get latest metadata:\n%s", err.Error())
	}
}

func Test_Metadata_Do_MissingChart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpResp := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(`{"error":"chart not found"}`)),
		Status:     "Not found",
		StatusCode: http.StatusNotFound,
	}

	mrt := mocks.NewMockRoundTripper(ctrl)
	mrt.EXPECT().RoundTrip(gomock.Any()).Return(httpResp, nil).Times(1)

	req, _ := request.NewRequest("localhost:1234")
	req.Transport = mrt
	mr := &MetadataReader{
		req: req,
	}

	_, err := mr.Do("foo")
	if err == nil {
		t.Errorf("Expected error but got none")
	}
}
