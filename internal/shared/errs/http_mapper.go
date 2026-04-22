package errs

import "errors"

func MapHTTP(err error) (int, string) {
	if err == nil {
		return 200, ""
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status, appErr.Message
	}

	return 500, err.Error()
}
