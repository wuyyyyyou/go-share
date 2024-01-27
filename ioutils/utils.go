package ioutils

import (
	"fmt"
	"io"
	"log"
	"os"
)

func CloseQuietly(cr ...io.Closer) {
	for _, c := range cr {
		if c != nil {
			if err := c.Close(); err != nil {
				log.Println(err)
			}
		}
	}
}

func CopyFile(dstFile, srcFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer CloseQuietly(src)

	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer CloseQuietly(dst)

	_, err = io.Copy(dst, src)
	return err
}

func ReaderToFile(src io.Reader, filename string) error {
	dst, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer CloseQuietly(dst)

	_, err = io.Copy(dst, src)
	return err
}

// FileExists 文件存在且不是文件夹
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func CreateDirsIfNotExists(dirName string) error {
	info, err := os.Stat(dirName)

	if err != nil {
		// 如果文件不存在则创建文件夹
		if os.IsNotExist(err) {
			return os.MkdirAll(dirName, os.ModePerm)
		}
	}

	if info.IsDir() {
		return nil
	}

	return fmt.Errorf("path exists but is not a directory")

}
