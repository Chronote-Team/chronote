package errs

type AppError struct {
	Status  int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func Validation(message string) error   { return &AppError{Status: 400, Message: message} }
func Unauthorized(message string) error { return &AppError{Status: 401, Message: message} }
func Forbidden(message string) error    { return &AppError{Status: 403, Message: message} }
func NotFound(message string) error     { return &AppError{Status: 404, Message: message} }
func Conflict(message string) error     { return &AppError{Status: 409, Message: message} }
func Internal(message string) error     { return &AppError{Status: 500, Message: message} }
func Degraded(message string) error     { return &AppError{Status: 207, Message: message} }
func Unavailable(message string) error  { return &AppError{Status: 503, Message: message} }
