package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

func Sha1hex(data []byte) string {
	h := sha1.New()
	h.Write(data)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func Md5hex(data []byte) string {
	h := md5.New()
	h.Write(data)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func Md5Reader(reader io.Reader) (string, int64, error) {
	h := md5.New()
	written, err := io.Copy(h, reader)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), written, nil
}

func Sha1Reader(reader io.Reader) (string, int64, error) {
	h := sha1.New()
	written, err := io.Copy(h, reader)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), written, nil
}

func Md5File(path string) (string, int64, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", 0, err
	}

	return Md5Reader(file)
}

func Sha1File(path string) (string, int64, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", 0, err
	}

	return Sha1Reader(file)
}
