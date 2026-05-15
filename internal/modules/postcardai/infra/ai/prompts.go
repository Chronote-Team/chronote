package ai

import _ "embed"

//go:embed prompts/image_understanding_v1.tmpl
var imageUnderstandingV1 string

//go:embed prompts/postcard_understanding_v1.tmpl
var postcardUnderstandingV1 string

func ImagePromptV1() string {
	return imageUnderstandingV1
}

func PostcardPromptV1() string {
	return postcardUnderstandingV1
}
