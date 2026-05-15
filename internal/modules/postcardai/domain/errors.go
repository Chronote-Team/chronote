package domain

import "fmt"

type ProviderErrorCode string

const (
	ErrorNone                ProviderErrorCode = ""
	ErrorProviderRefused     ProviderErrorCode = "provider_refused"
	ErrorProviderUnavailable ProviderErrorCode = "provider_unavailable"
	ErrorProviderTimeout     ProviderErrorCode = "provider_timeout"
	ErrorMalformedOutput     ProviderErrorCode = "malformed_output"
	ErrorPermanentInput      ProviderErrorCode = "permanent_input_failure"
	ErrorStaleVersion        ProviderErrorCode = "stale_version"
)

type ProviderError struct {
	Code ProviderErrorCode
	Err  error
}

func (e ProviderError) Error() string {
	if e.Err == nil {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

func (e ProviderError) Unwrap() error {
	return e.Err
}

func ErrorCode(err error) ProviderErrorCode {
	if err == nil {
		return ErrorNone
	}
	if providerErr, ok := err.(ProviderError); ok {
		return providerErr.Code
	}
	return ErrorProviderUnavailable
}

func IsPermanent(code ProviderErrorCode) bool {
	return code == ErrorProviderRefused || code == ErrorPermanentInput
}
