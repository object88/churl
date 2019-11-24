package churl

import (
	"encoding/json"
	"fmt"

	"github.com/object88/churl/internal/request"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/repo"
)

type MetadataReader struct {
	req *request.Request
}

func NewMetadataReader(sourceport string) (*MetadataReader, error) {
	baseURL := fmt.Sprintf("http://localhost:%s", sourceport)
	req, err := request.NewRequest(baseURL)
	if err != nil {
		return nil, err
	}

	m := &MetadataReader{
		req: req,
	}
	return m, nil
}

func (m *MetadataReader) Do(chartpath string) (*repo.ChartVersion, error) {
	query := fmt.Sprintf("api/charts/%s", chartpath)
	rc, ok, err := m.req.ProcessGet(query)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query for chart")
	}
	defer rc.Close()

	dec := json.NewDecoder(rc)

	if !ok {
		aerr := &ApiError{}
		dec.Decode(aerr)
		return nil, aerr
	}

	cvs := []*repo.ChartVersion{}
	err = dec.Decode(&cvs)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to decode")
	}

	if cvs == nil || len(cvs) == 0 {
		return nil, nil
	}

	cv := cvs[0]

	return cv, nil
}
