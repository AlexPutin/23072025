package utils

import (
	"fmt"
	"net/url"
	"path"
)

func GetUrlExtension(fileUrl string) (string, error) {
	parsedURL, err := url.Parse(fileUrl)
	if err != nil {
		return "", fmt.Errorf("invalid url")
	}

	urlPath := parsedURL.Path
	ext := path.Ext(urlPath)
	return ext, nil
}
