package infra

type ImageProcessor struct{}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

func (ImageProcessor) ImageMetadata(filename string, data []byte) (int, int) {
	return 0, 0
}
