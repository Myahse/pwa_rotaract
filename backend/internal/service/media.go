package service

import (
	"errors"
	"mime"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidImageType = errors.New("image must be jpeg, png, or webp")
	ErrImageTooLarge    = errors.New("image file is too large")
)

func imageExtension(filename, contentType string) (string, error) {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = mime.TypeByExtension(filepath.Ext(filename))
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	}

	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", ErrInvalidImageType
	}
}
