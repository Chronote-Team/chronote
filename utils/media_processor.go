package utils

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"strings"
)

var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

var videoExtensions = map[string]bool{
	".mp4":  true,
	".mov":  true,
	".webm": true,
	".avi":  true,
	".mkv":  true,
}

var audioExtensions = map[string]bool{
	".mp3": true,
	".wav": true,
	".aac": true,
	".m4a": true,
	".ogg": true,
}

func DetectMediaType(filename, contentType string) (string, error) {
	contentType = strings.ToLower(contentType)
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return "image", nil
	case strings.HasPrefix(contentType, "video/"):
		return "video", nil
	case strings.HasPrefix(contentType, "audio/"):
		return "audio", nil
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if imageExtensions[ext] {
		return "image", nil
	}
	if videoExtensions[ext] {
		return "video", nil
	}
	if audioExtensions[ext] {
		return "audio", nil
	}
	return "", errors.New("不支持的媒体类型")
}

func GetImageDimensions(data []byte) (int, int, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func GenerateImageThumbnail(data []byte, maxSize int) ([]byte, error) {
	if maxSize <= 0 {
		return data, nil
	}
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	thumbnail := resizeImage(img, maxSize)
	buf := new(bytes.Buffer)
	switch format {
	case "jpeg":
		err = jpeg.Encode(buf, thumbnail, &jpeg.Options{Quality: 80})
	case "png":
		err = png.Encode(buf, thumbnail)
	case "gif":
		err = gif.Encode(buf, thumbnail, nil)
	default:
		return nil, errors.New("不支持的图片格式")
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func resizeImage(src image.Image, maxSize int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= maxSize && height <= maxSize {
		return src
	}
	maxDimension := width
	if height > width {
		maxDimension = height
	}
	scale := float64(maxSize) / float64(maxDimension)
	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := x * width / newWidth
			srcY := y * height / newHeight
			dst.Set(x, y, src.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}
	return dst
}
