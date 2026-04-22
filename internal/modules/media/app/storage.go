package app

type Storage interface {
	Upload(key, filename string, data []byte, contentType string) (string, error)
	Delete(key string) error
}
