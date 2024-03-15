package receiver

import "errors"

var (
	ErrSendingReq     = errors.New("error sending request")
	ErrStatusIsNot200 = errors.New("error wrong status code")
)
