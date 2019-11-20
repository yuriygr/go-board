package uploader

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"

	"io"
	"mime/multipart"
	"os"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
}

// ImageDimensions - Структура с параметрами изображения
type ImageDimensions struct {
	Width  int
	Height int
	Size   int64
	Md5    string
}

// WriteImageFile - Write file to path
func WriteImageFile(path string, file multipart.File) (*ImageDimensions, error) {
	tempFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		return nil, err
	}

	return ImageDimensionsByPath(path)
}

// ImageDimensionsByPath - Get file dimensions
func ImageDimensionsByPath(path string) (*ImageDimensions, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var fileBuffer bytes.Buffer
	tee := io.TeeReader(file, &fileBuffer)
	hash := md5.New()
	fileReaderBuffer, _ := ioutil.ReadAll(&fileBuffer)
	_, err = io.Copy(hash, bytes.NewReader(fileReaderBuffer))
	if err != nil {
		return nil, err
	}
	md5 := fmt.Sprintf("%x", hash.Sum(nil))

	fileReader, _ := ioutil.ReadAll(tee)
	image, _, err := image.DecodeConfig(bytes.NewReader(fileReader))
	if err != nil {
		return nil, err
	}

	return &ImageDimensions{image.Width, image.Height, stat.Size(), md5}, nil
}

// ImageDimensionsByFile - Get file dimensions
func ImageDimensionsByFile(file *os.File) (*ImageDimensions, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	return &ImageDimensions{image.Width, image.Height, stat.Size(), ""}, nil
}
