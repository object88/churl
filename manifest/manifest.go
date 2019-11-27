package manifest

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sort"

	"github.com/pkg/errors"
)

const (
	museumsKey string = "museums"
	currentKey string = "current"
)

// Manifest describes the configuration for the churl tool
type Manifest struct {
	Museums map[string]*ChartMuseum

	current string

	f *os.File
}

// Init creates a new manifest instance and creates a new file at `target`.  If
// `target` already exists, func fails.  File is created but has no contents
// until Save is called.
func Init(target string) (*Manifest, error) {
	m := New()

	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open manifest file '%s'", target)
	}

	m.f = f

	return m, nil
}

// New creates a new manifest instance
func New() *Manifest {
	return &Manifest{
		Museums: map[string]*ChartMuseum{},
	}
}

// Open creates a Manifest instance from the JSON content from the `r`
// parameter
func Open(r io.Reader) (*Manifest, error) {
	m := &Manifest{}

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(m); err != nil {
		return nil, errors.Wrapf(err, "Failed to decode the manifest")
	}

	return m, nil
}

// OpenFromFile creates a Manifest instance from the contents of the JSON-
// encoded contents of the file at `manifestFilepath`.  An open reference to
// the file is kept with the instance, so the caller is responsible for calling
// `Close`.
func OpenFromFile(manifestFilepath string) (*Manifest, error) {
	f, err := os.Open(manifestFilepath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open manifest file '%s'", manifestFilepath)
	}

	m, err := Open(f)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open manifest from file '%s'", manifestFilepath)
	}

	m.f = f

	return m, nil
}

// Current returns the current chart museum, or nil
func (m *Manifest) Current() *ChartMuseum {
	if m == nil || m.current == "" {
		return nil
	}

	return m.Museums[m.current]
}

// Save write the manifest file to disk, if it was opened with `OpenFromFile`
// or created with `Init`
func (m *Manifest) Save() error {
	if m == nil {
		return errors.Errorf("manifest pointer reciever is nil; cannot save")
	}
	if m.f == nil {
		return errors.Errorf("no file is open, cannot save")
	}

	_, err := m.f.Seek(0, 0)
	if err != nil {
		return errors.Wrapf(err, "internal error: failed to keep to beginning of file")
	}

	n, err := m.WriteTo(m.f)
	if err != nil {
		return errors.Wrapf(err, "failed to save manifest")
	}

	err = m.f.Truncate(n)
	if err != nil {
		return errors.Wrapf(err, "internal error: failed to truncate manifest to %d bytes", n)
	}

	return nil
}

// Close satisfies the io.Closer interface
func (m *Manifest) Close() error {
	if m.f != nil {
		defer func() {
			m.f = nil
		}()

		err := m.f.Close()
		if err != nil {
			return errors.Wrapf(err, "failed to close manifest file")
		}
	}

	return nil
}

// MarshalJSON satisfies the encoding/json.Marshaler interface
func (m *Manifest) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`{"`)
	buf.WriteString(museumsKey)
	buf.WriteString(`":[`)

	names := make([]string, len(m.Museums))
	offset := 0
	for key := range m.Museums {
		names[offset] = key
		offset++
	}
	sort.Strings(names)

	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")

	for k, v := range names {
		if k != 0 {
			buf.WriteRune(',')
		}
		err := enc.Encode(intermediateMuseum{
			name:        v,
			ChartMuseum: m.Museums[v],
		})
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to encode intermediateMuseum")
		}
	}

	buf.WriteString(`], "current": "`)
	buf.WriteString(m.current)
	buf.WriteString(`"}`)

	return buf.Bytes(), nil
}

// UnmarshalJSON satisfies the encoding/json.Unmarshaler interface
func (m *Manifest) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal top-level manifest")
	}

	m.Museums = map[string]*ChartMuseum{}

	r, ok := objMap[museumsKey]
	if !ok {
		return errors.Wrapf(err, "Must have '%s' key", museumsKey)
	}

	var data []*intermediateMuseum
	err = json.Unmarshal(*r, &data)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal '%s' value into array of *intermediateMuseum", museumsKey)
	}

	for _, v := range data {
		m.Museums[v.name] = v.ChartMuseum
	}

	// Read in the "current" key to set the current museum
	r, ok = objMap[currentKey]
	if !ok {
		return errors.Errorf("Must have '%s' key", currentKey)
	}
	err = json.Unmarshal(*r, &m.current)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal '%s' value into string", currentKey)
	}

	if _, ok = m.Museums[m.current]; !ok {
		return errors.Errorf("Manifest JSON contains invalid 'current' museum '%s'", m.current)
	}

	return nil
}

// WriteTo satisfies the io.WriterTo interface
func (m *Manifest) WriteTo(w io.Writer) (int64, error) {
	wc := writeCounter{
		w: w,
	}
	enc := json.NewEncoder(&wc)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return int64(wc.count), errors.Wrapf(err, "Failed to encode the manifest")
	}
	return int64(wc.count), nil
}
