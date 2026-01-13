package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(url string) (localPath string, err error) {
	// Faz o download
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: status %s", resp.Status)
	}

	// Extrai o nome do arquivo da URL (fallback para temp)
	filename := filepath.Base(resp.Request.URL.Path)
	if filename == "." || filename == "/" {
		filename = "downloaded_file"
	}

	// Cria arquivo temporário
	tmpFile, err := os.CreateTemp("", "download-*-"+filename)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			os.Remove(tmpFile.Name())
		}
	}()

	// Copia o conteúdo
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", err
	}

	// Garante flush no disco
	if err = tmpFile.Close(); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

type FileInfo struct {
	Size     int64
	MimeType string
	TempPath string
}

func GetFileInfo(filePath string) (*FileInfo, error) {
	var info FileInfo
	info.TempPath, _ = filepath.Abs(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo %s: %v", filePath, err)
	}
	defer file.Close()

	// Obter tamanho do arquivo
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informações do arquivo %s: %v", filePath, err)
	}
	info.Size = stat.Size()

	// Detectar MIME type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("erro ao ler arquivo %s para detecção de MIME type: %v", filePath, err)
	}
	info.MimeType = http.DetectContentType(buffer[:n])

	return &info, nil
}
