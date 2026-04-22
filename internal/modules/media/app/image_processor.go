package app

type ImageProcessor interface {
	ImageMetadata(filename string, data []byte) (width, height int)
}
