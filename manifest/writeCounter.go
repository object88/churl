package manifest

import "io"

type writeCounter struct {
	count int
	w     io.Writer
}

func (wc *writeCounter) Write(b []byte) (int, error) {
	n, err := wc.w.Write(b)
	wc.count += n
	return n, err
}
