package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// MaxURLDownloadTimeout representa o timeout máximo para download de URL (30 segundos)
	MaxURLDownloadTimeout = 30 * time.Second
)

// DownloadFileFromURL baixa um arquivo de uma URL e retorna informações do arquivo
// Retorna o mesmo formato de Base64FileInfo para compatibilidade
func DownloadFileFromURL(url string) (*Base64FileInfo, error) {
	if url == "" {
		return nil, errors.New("URL não pode estar vazia")
	}

	// Criar cliente HTTP com timeout
	client := &http.Client{
		Timeout: MaxURLDownloadTimeout,
	}

	// Fazer requisição GET
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer download da URL: %v", err)
	}
	defer resp.Body.Close()

	// Verificar status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao baixar arquivo: status code %d", resp.StatusCode)
	}

	// Ler conteúdo do arquivo
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler conteúdo do arquivo: %v", err)
	}

	// Verificar tamanho após download
	if len(fileData) > MaxFileSize {
		return nil, fmt.Errorf("tamanho do arquivo após download excede o limite de %.1f MB", MaxFileSize/(1024*1024))
	}

	// Detectar MIME type usando amostra adequada
	sampleSize := len(fileData)
	if sampleSize > 512 {
		sampleSize = 512
	}
	mimeType := http.DetectContentType(fileData[:sampleSize])

	// Validar MIME type
	if err := ValidateMimeType(mimeType); err != nil {
		return nil, fmt.Errorf("tipo de arquivo não suportado: %v", err)
	}

	// Criar arquivo temporário
	tempFile, err := os.CreateTemp("", "docsigner_url_*")
	if err != nil {
		return nil, fmt.Errorf("erro ao criar arquivo temporário: %v", err)
	}

	// Escrever dados no arquivo temporário
	if _, err := tempFile.Write(fileData); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("erro ao escrever arquivo temporário: %v", err)
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("erro ao fechar arquivo temporário: %v", err)
	}

	return &Base64FileInfo{
		DecodedData: fileData,
		MimeType:    mimeType,
		Size:        int64(len(fileData)),
		TempPath:    tempFile.Name(),
	}, nil
}



