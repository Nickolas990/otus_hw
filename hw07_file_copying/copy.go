package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	//nolint:depguard
	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	fmt.Printf("Copying from %s to %s with offset %d and limit %d\n", fromPath, toPath, offset, limit)

	srcFile, err := os.Open(fromPath)
	if err != nil {
		log.Printf("failed to open source file: %v", err)
		return err
	}
	defer srcFile.Close()

	// Проверка размера файла
	fileInfo, err := srcFile.Stat()
	if err != nil {
		log.Printf("failed to stat source file: %v", err)
		return err
	}

	// Проверка для специальных файлов
	if !fileInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	if fileInfo.Size() < offset {
		return ErrOffsetExceedsFileSize
	}

	// Проверка, являются ли файлы одинаковыми
	destInfo, err := os.Stat(toPath)
	if err == nil && os.SameFile(fileInfo, destInfo) {
		log.Printf("Source and destination files are the same. No copy needed.")
		return nil
	}

	// Создание/открытие целевого файла
	destFile, err := os.Create(toPath)
	if err != nil {
		log.Printf("failed to create destination file: %v", err)
		return err
	}
	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {
			log.Printf("failed to close destination file: %v", err)
		}
	}(destFile)

	// Установка смещения
	_, err = srcFile.Seek(offset, io.SeekStart)
	if err != nil {
		log.Printf("failed to seek in source file: %v", err)
		return err
	}

	// Определение лимита
	if limit == 0 || limit > fileInfo.Size()-offset {
		limit = fileInfo.Size() - offset
	}

	// Инициализация прогресс-бара
	bar := pb.Full.Start64(limit)
	bar.SetWidth(40)
	bar.Set(pb.Bytes, true) // Отображение в байтах
	bar.SetTemplateString(`{{bar . }} {{percent . }} {{counters . }}`)
	defer bar.Finish()

	// Обертка для отслеживания прогресса
	reader := bar.NewProxyReader(io.LimitReader(srcFile, limit))

	// Копирование данных
	bytesCopied, err := io.CopyN(destFile, reader, limit)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Printf("failed to copy data: %v", err)
		return err
	}

	fmt.Printf("Successfully copied %d bytes from %s to %s\n", bytesCopied, fromPath, toPath)
	return nil
}
