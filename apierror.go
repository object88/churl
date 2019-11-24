package churl

type ApiError struct {
	Err string `json:"error"`
}

func (aerr *ApiError) Error() string {
	return aerr.Err
}
