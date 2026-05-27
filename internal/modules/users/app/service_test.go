package app

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
)

func TestRegisterNormalizesEmailAndDefaultsDisplayName(t *testing.T) {
	service := NewService(nil, nil)

	user, err := service.Register(RegisterInput{
		Username: "tester",
		Email:    "TEST@EXAMPLE.COM",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Fatalf("expected lowercased email, got %q", user.Email)
	}
	if user.DisplayName != "tester" {
		t.Fatalf("expected default display name, got %q", user.DisplayName)
	}
}

func TestUploadAvatarStoresFileAndUpdatesUserAvatar(t *testing.T) {
	storage := &recordingAvatarStorage{url: "https://media.example.com/avatars/1/avatar.png"}
	service := NewService(nil, nil)
	service.SetAvatarStorage(storage)

	user, err := service.Register(RegisterInput{
		Username: "tester",
		Email:    "tester@example.com",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	file := multipartFileHeader(t, "avatar.png", "image/png", []byte("fake-png-data"))

	url, err := service.UploadAvatar(user.ID, file)
	if err != nil {
		t.Fatalf("UploadAvatar returned error: %v", err)
	}

	if url != storage.url {
		t.Fatalf("expected storage URL %q, got %q", storage.url, url)
	}
	if storage.filename != "avatar.png" {
		t.Fatalf("expected filename avatar.png, got %q", storage.filename)
	}
	if storage.contentType != "image/png" {
		t.Fatalf("expected image/png content type, got %q", storage.contentType)
	}
	if !strings.HasPrefix(storage.key, "avatars/1/") {
		t.Fatalf("expected avatar key under user folder, got %q", storage.key)
	}
	if string(storage.data) != "fake-png-data" {
		t.Fatalf("unexpected uploaded bytes: %q", string(storage.data))
	}

	updated, err := service.GetUserInfo(user.ID)
	if err != nil {
		t.Fatalf("GetUserInfo returned error: %v", err)
	}
	if updated.Avatar != storage.url {
		t.Fatalf("expected saved avatar %q, got %q", storage.url, updated.Avatar)
	}
}

func TestUploadAvatarRejectsNonImageFile(t *testing.T) {
	storage := &recordingAvatarStorage{url: "https://media.example.com/avatars/1/avatar.txt"}
	service := NewService(nil, nil)
	service.SetAvatarStorage(storage)

	user, err := service.Register(RegisterInput{
		Username: "tester",
		Email:    "tester@example.com",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	file := multipartFileHeader(t, "avatar.txt", "text/plain", []byte("not-image"))

	if _, err := service.UploadAvatar(user.ID, file); err == nil || !strings.Contains(err.Error(), "头像文件类型无效") {
		t.Fatalf("expected invalid avatar type error, got %v", err)
	}
	if storage.key != "" {
		t.Fatalf("storage should not be called for invalid avatar, got key %q", storage.key)
	}
}

func TestUploadAvatarAcceptsGenericMultipartContentTypeForImageExtension(t *testing.T) {
	storage := &recordingAvatarStorage{url: "https://media.example.com/avatars/1/avatar.png"}
	service := NewService(nil, nil)
	service.SetAvatarStorage(storage)

	user, err := service.Register(RegisterInput{
		Username: "tester",
		Email:    "tester@example.com",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	file := multipartFileHeader(t, "avatar.png", "application/octet-stream", []byte("fake-png-data"))

	if _, err := service.UploadAvatar(user.ID, file); err != nil {
		t.Fatalf("UploadAvatar returned error for generic multipart content type: %v", err)
	}
	if storage.key == "" {
		t.Fatalf("expected storage upload for image extension with generic content type")
	}
}

type recordingAvatarStorage struct {
	url         string
	key         string
	filename    string
	data        []byte
	contentType string
}

func (s *recordingAvatarStorage) Upload(key, filename string, data []byte, contentType string) (string, error) {
	s.key = key
	s.filename = filename
	s.data = append([]byte(nil), data...)
	s.contentType = contentType
	return s.url, nil
}

func multipartFileHeader(t *testing.T, filename, contentType string, data []byte) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="avatar"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	reader := multipart.NewReader(bytes.NewReader(body.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(int64(body.Len()))
	if err != nil {
		t.Fatalf("read multipart form: %v", err)
	}
	files := form.File["avatar"]
	if len(files) != 1 {
		t.Fatalf("expected one avatar file, got %d", len(files))
	}
	return files[0]
}
