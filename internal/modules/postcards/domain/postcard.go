package domain

import (
	"encoding/json"
	"time"

	mediadomain "chronote-refactor/internal/modules/media/domain"
)

type Author struct {
	ID          uint
	Username    string
	DisplayName string
	Avatar      string
}

type Postcard struct {
	ID         uint
	Title      string
	Content    json.RawMessage
	Visibility string
	AuthorID   uint
	Author     *Author
	Medias     []mediadomain.Media
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
