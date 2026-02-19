package services

import "errors"

var (
	ErrorInvalidCredentials = errors.New("error passwords didn't match")
	ErrorInvalidOrderNumber = errors.New("error the number is incorrect")
)
