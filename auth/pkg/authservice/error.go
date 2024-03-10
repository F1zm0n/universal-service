package authservice

import "errors"

var (
	ErrWrongEmailFmt   = errors.New("wrong email format")
	ErrWrongPassFmt    = errors.New("password should be minimum 8 length long")
	ErrInsertingUser   = errors.New("error inserting user")
	ErrInvalidCreds    = errors.New("invalid credentials")
	ErrGeneratingToken = errors.New("error generating jwt token")
)
