package forwarder

type nilwriter struct{}

func (_ *nilwriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
