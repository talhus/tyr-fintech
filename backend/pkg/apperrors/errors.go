package apperrors

import "net/http"

type AppError struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Msg
}

func New(code int, message string) *AppError {
	return &AppError{
		Code: code,
		Msg:  message,
	}
}

var (
	ErrUserNotFound             = New(http.StatusNotFound, "User profile could not be found.")
	ErrUserAlreadyExists        = New(http.StatusBadRequest, "An account with this email already exists.")
	ErrInvalidCredentials       = New(http.StatusUnauthorized, "Invalid email or password. Please try again.")
	ErrInsufficientBalance      = New(http.StatusBadRequest, "Insufficient funds to complete this transaction.")
	ErrUserNotHaveWallet        = New(http.StatusBadRequest, "No wallet associated with this account.")
	ErrInvalidPassword          = New(http.StatusBadRequest, "The password you entered is incorrect.")
	ErrWalletAlreadyExists      = New(http.StatusBadRequest, "A wallet with this currency already exists.")
	ErrTransactionNotFound      = New(http.StatusNotFound, "Requested transaction could not be located.")
	ErrTransactionStatusInvalid = New(http.StatusBadRequest, "Invalid transaction status payload.")
	ErrInternalServer           = New(http.StatusInternalServerError, "An unexpected error occurred. Please try again later.")
)
