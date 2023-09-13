package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func getProgressbar(progressPercent, progressBarLen int) (progressBar string) {
	i := 0
	for ; i < progressPercent/(100/progressBarLen); i++ {
		progressBar += "▰"
	}
	for ; i < progressBarLen; i++ {
		progressBar += "▱"
	}
	progressBar += " " + fmt.Sprint(progressPercent) + "%"
	return
}

func fileNameWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func copyFile(dst, src string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	buf := make([]byte, 1048576)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}
