package domain

import "errors"

var (
	ErrorUUIDGenerate       = errors.New("error during uuid generation")
	ErrorUserAlreadyExists  = errors.New("the user already exists")
	ErrorDBWriting          = errors.New("error when creating or updating data in the database")
	ErrorUserNotFound       = errors.New("no such user was found")
	ErrorDBReading          = errors.New("error while reading the database")
	ErrorOrderAlreadyExists = errors.New("the order has already been uploaded by this user")
	ErrorOrderConflict      = errors.New("the order was uploaded by another user")
	ErrorDataScan           = errors.New("error during data scanning")
	ErrorRowsIteration      = errors.New("error during rows iteration")
	ErrorTransactionStart   = errors.New("error when starting a transaction")
	ErrorInsufficientFunds  = errors.New("there are insufficient funds in the account")
)

var (
	ErrorInvalidCredentials = errors.New("error passwords didn't match")
	ErrorInvalidOrderNumber = errors.New("error the number is incorrect")
)
