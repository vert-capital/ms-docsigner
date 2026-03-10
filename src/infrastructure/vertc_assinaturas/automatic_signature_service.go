package vertc_assinaturas

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// AutomaticSignatureCheckResult representa o resultado da validação por e-mail.
type AutomaticSignatureCheckResult struct {
	HasSignedTerm   bool
	PermissionFound bool
	PermissionID    string
	ContractStatus  string
	IsActive        *bool
}

// AutomaticSignatureCreateResult representa o resultado da criação de uma permissão.
type AutomaticSignatureCreateResult struct {
	PermissionID      string
	EnvelopeID        string
	ContractStatus    string
	IsActive          *bool
	NotificationSent  bool
	NotificationError *string
	UserCreated       bool
	UserExisted       bool
}

type vertSignUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type createVertSignUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
	IsActive bool   `json:"isActive"`
}

type automaticSignatureUserResponse struct {
	Email string `json:"email"`
}

type automaticSignaturePermissionResponse struct {
	ID string `json:"id"`

	EnvelopeID      string `json:"envelopeId"`
	EnvelopeIDSnake string `json:"envelope_id"`

	RecipientUser      automaticSignatureUserResponse `json:"recipientUser"`
	RecipientUserSnake automaticSignatureUserResponse `json:"recipient_user"`

	ContractStatus      string `json:"contractStatus"`
	ContractStatusSnake string `json:"contract_status"`

	IsActive      *bool `json:"isActive"`
	IsActiveSnake *bool `json:"is_active"`

	RevokedAt      *time.Time `json:"revokedAt"`
	RevokedAtSnake *time.Time `json:"revoked_at"`

	UpdatedAt      *time.Time `json:"updatedAt"`
	UpdatedAtSnake *time.Time `json:"updated_at"`
}

func (p automaticSignaturePermissionResponse) recipientEmail() string {
	if p.RecipientUser.Email != "" {
		return p.RecipientUser.Email
	}
	return p.RecipientUserSnake.Email
}

func (p automaticSignaturePermissionResponse) contractStatus() string {
	if p.ContractStatus != "" {
		return p.ContractStatus
	}
	return p.ContractStatusSnake
}

func (p automaticSignaturePermissionResponse) envelopeID() string {
	if p.EnvelopeID != "" {
		return p.EnvelopeID
	}
	return p.EnvelopeIDSnake
}

func (p automaticSignaturePermissionResponse) active() bool {
	if p.IsActive != nil {
		return *p.IsActive
	}
	if p.IsActiveSnake != nil {
		return *p.IsActiveSnake
	}
	return false
}

func (p automaticSignaturePermissionResponse) revoked() bool {
	return p.RevokedAt != nil || p.RevokedAtSnake != nil
}

func (p automaticSignaturePermissionResponse) updatedAtOrZero() time.Time {
	if p.UpdatedAt != nil {
		return *p.UpdatedAt
	}
	if p.UpdatedAtSnake != nil {
		return *p.UpdatedAtSnake
	}
	return time.Time{}
}

// AutomaticSignatureService consulta e cria permissões de assinatura automática na VertSign.
type AutomaticSignatureService struct {
	client *VertcAssinaturasClient
	logger *logrus.Logger
}

// NewAutomaticSignatureService cria uma nova instância do serviço de assinatura automática.
func NewAutomaticSignatureService(client *VertcAssinaturasClient, logger *logrus.Logger) *AutomaticSignatureService {
	return &AutomaticSignatureService{
		client: client,
		logger: logger,
	}
}

// CheckSignedTermByEmail verifica se o e-mail informado possui termo ativo e assinado.
func (s *AutomaticSignatureService) CheckSignedTermByEmail(ctx context.Context, signerEmail string) (*AutomaticSignatureCheckResult, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(signerEmail))
	if normalizedEmail == "" {
		return nil, fmt.Errorf("signer email is required")
	}

	resp, err := s.client.Get(ctx, "/api/v1/automatic-signature")
	if err != nil {
		return nil, fmt.Errorf("failed to call automatic-signature endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read automatic-signature response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &VertcAssinaturasError{
			Type:       s.client.categorizeHTTPError(resp.StatusCode),
			Message:    fmt.Sprintf("automatic-signature failed with status %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	permissions, err := parseAutomaticSignaturePermissions(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse automatic-signature response: %w", err)
	}

	var latestMatch *automaticSignaturePermissionResponse

	for i := range permissions {
		permission := permissions[i]
		recipientEmail := strings.ToLower(strings.TrimSpace(permission.recipientEmail()))

		if recipientEmail != normalizedEmail {
			continue
		}

		if latestMatch == nil || permission.updatedAtOrZero().After(latestMatch.updatedAtOrZero()) {
			latestMatch = &permissions[i]
		}

		status := strings.ToLower(strings.TrimSpace(permission.contractStatus()))
		isActive := permission.active()

		if status == "signed" && isActive && !permission.revoked() {
			active := isActive
			return &AutomaticSignatureCheckResult{
				HasSignedTerm:   true,
				PermissionFound: true,
				PermissionID:    permission.ID,
				ContractStatus:  permission.contractStatus(),
				IsActive:        &active,
			}, nil
		}
	}

	// Se existe permissão para o e-mail, retorna o último status observado.
	if latestMatch != nil {
		active := latestMatch.active()
		return &AutomaticSignatureCheckResult{
			HasSignedTerm:   false,
			PermissionFound: true,
			PermissionID:    latestMatch.ID,
			ContractStatus:  latestMatch.contractStatus(),
			IsActive:        &active,
		}, nil
	}

	return &AutomaticSignatureCheckResult{
		HasSignedTerm:   false,
		PermissionFound: false,
	}, nil
}

// CreateTermEnsuringUser cria uma permissão de assinatura automática, garantindo antes que o destinatário exista como usuário.
func (s *AutomaticSignatureService) CreateTermEnsuringUser(ctx context.Context, signerEmail, signerName string) (*AutomaticSignatureCreateResult, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(signerEmail))
	if normalizedEmail == "" {
		return nil, fmt.Errorf("signer email is required")
	}

	userCreated := false
	userExisted := false

	exists, err := s.userExistsByEmail(ctx, normalizedEmail)
	if err != nil {
		s.logger.WithError(err).WithField("signer_email", normalizedEmail).Warn("Failed to check if user exists in vert-sign. Falling back to create-on-demand strategy")
	} else {
		userExisted = exists
		if !exists {
			created, createErr := s.createUserIfMissing(ctx, normalizedEmail, signerName)
			if createErr != nil {
				return nil, fmt.Errorf("failed to create user before automatic-signature creation: %w", createErr)
			}
			userCreated = created
			userExisted = !created
		}
	}

	permission, err := s.createAutomaticSignaturePermission(ctx, normalizedEmail)
	if err != nil && isRecipientUserNotFoundError(err) {
		created, createErr := s.createUserIfMissing(ctx, normalizedEmail, signerName)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create missing recipient user: %w", createErr)
		}

		if created {
			userCreated = true
			userExisted = false
		} else {
			userExisted = true
		}

		permission, err = s.createAutomaticSignaturePermission(ctx, normalizedEmail)
	}

	if err != nil {
		return nil, err
	}

	envelopeID := permission.envelopeID()
	if envelopeID == "" {
		return nil, fmt.Errorf("automatic-signature response does not contain envelope_id")
	}

	pdfFileName := fmt.Sprintf("termo_assinatura_automatica_%s.pdf", time.Now().UTC().Format("20060102_150405"))
	pdfBytes, err := generateAutomaticSignatureTermPDF(normalizedEmail, signerName, envelopeID, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to generate automatic-signature term pdf: %w", err)
	}

	if err := s.uploadAutomaticSignatureTermDocument(ctx, envelopeID, pdfFileName, pdfBytes); err != nil {
		return nil, fmt.Errorf("failed to upload automatic-signature term document: %w", err)
	}

	notificationSent := false
	var notificationError *string

	if err := s.sendAutomaticSignatureEnvelope(ctx, envelopeID); err != nil {
		if isEnvelopeMissingDocumentsForSendError(err) {
			message := "Envelope do termo criado sem documentos; notificação será enviada após upload do termo e envio do envelope."
			notificationError = &message

			s.logger.WithError(err).WithFields(logrus.Fields{
				"envelope_id": envelopeID,
				"email":       normalizedEmail,
			}).Warn("Automatic-signature envelope created without documents; notification was not sent yet")
		} else {
			return nil, fmt.Errorf("failed to send automatic-signature envelope notification: %w", err)
		}
	} else {
		notificationSent = true
	}

	// Se a permissão foi criada com sucesso, o usuário necessariamente existe no vert-sign.
	if !userCreated && !userExisted {
		userExisted = true
	}

	active := permission.active()
	return &AutomaticSignatureCreateResult{
		PermissionID:      permission.ID,
		EnvelopeID:        envelopeID,
		ContractStatus:    permission.contractStatus(),
		IsActive:          &active,
		NotificationSent:  notificationSent,
		NotificationError: notificationError,
		UserCreated:       userCreated,
		UserExisted:       userExisted,
	}, nil
}

func (s *AutomaticSignatureService) userExistsByEmail(ctx context.Context, email string) (bool, error) {
	resp, err := s.client.Get(ctx, "/api/v1/users")
	if err != nil {
		return false, fmt.Errorf("failed to call users endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read users response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, &VertcAssinaturasError{
			Type:       s.client.categorizeHTTPError(resp.StatusCode),
			Message:    fmt.Sprintf("users lookup failed with status %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	var users []vertSignUserResponse
	if err := json.Unmarshal(body, &users); err != nil {
		return false, fmt.Errorf("failed to parse users response: %w", err)
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	matches := 0
	for _, user := range users {
		if strings.ToLower(strings.TrimSpace(user.Email)) == normalizedEmail {
			matches++
		}
	}

	if matches > 1 {
		return false, fmt.Errorf("multiple users found for email %s", normalizedEmail)
	}

	return matches == 1, nil
}

func (s *AutomaticSignatureService) createUserIfMissing(ctx context.Context, email, name string) (bool, error) {
	password, err := generateRandomPassword()
	if err != nil {
		return false, fmt.Errorf("failed to generate random password: %w", err)
	}

	userName := strings.TrimSpace(name)
	if userName == "" {
		userName = defaultUserNameFromEmail(email)
	}

	request := createVertSignUserRequest{
		Email:    email,
		Name:     userName,
		Password: password,
		Role:     "user",
		IsActive: true,
	}

	resp, err := s.client.Post(ctx, "/api/v1/users", request, "")
	if err != nil {
		if isUserAlreadyExistsError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to create user in vert-sign: %w", err)
	}
	defer resp.Body.Close()

	return true, nil
}

func (s *AutomaticSignatureService) createAutomaticSignaturePermission(ctx context.Context, email string) (*automaticSignaturePermissionResponse, error) {
	request := map[string]string{"email": email}

	resp, err := s.client.Post(ctx, "/api/v1/automatic-signature", request, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create automatic-signature permission: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read automatic-signature creation response: %w", err)
	}

	permission, err := parseAutomaticSignaturePermission(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse automatic-signature creation response: %w", err)
	}

	return permission, nil
}

func (s *AutomaticSignatureService) sendAutomaticSignatureEnvelope(ctx context.Context, envelopeID string) error {
	endpoint := fmt.Sprintf("/api/v1/envelopes/%s/send", envelopeID)
	payload := map[string]string{}

	resp, err := s.client.Post(ctx, endpoint, payload, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (s *AutomaticSignatureService) uploadAutomaticSignatureTermDocument(ctx context.Context, envelopeID, fileName string, fileContent []byte) error {
	endpoint := fmt.Sprintf("/api/v1/documents/%s", envelopeID)
	resp, err := s.client.PostMultipartFile(ctx, endpoint, "files", fileName, fileContent)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func generateAutomaticSignatureTermPDF(signerEmail, signerName, envelopeID string, generatedAt time.Time) ([]byte, error) {
	var lines []string

	name := strings.TrimSpace(signerName)
	if name == "" {
		name = defaultUserNameFromEmail(signerEmail)
	}

	lines = append(lines,
		"TERMO DE AUTORIZACAO DE ASSINATURA AUTOMATICA",
		"",
		"Documento gerado dinamicamente pela integracao ms-docsigner.",
		"",
		fmt.Sprintf("Signatario destinatario: %s", name),
		fmt.Sprintf("Email destinatario: %s", signerEmail),
		fmt.Sprintf("Envelope ID: %s", envelopeID),
		fmt.Sprintf("Data de geracao (UTC): %s", generatedAt.Format("2006-01-02 15:04:05")),
		"",
		"Declaracao:",
		"O signatario destinatario, ao assinar este documento no VertSign,",
		"autoriza a execucao de assinaturas automaticas conforme fluxo acordado",
		"entre as partes no ambiente Vert.",
		"",
		"Este termo integra o registro de evidencias da plataforma.",
	)

	return buildSimplePDF(lines)
}

func buildSimplePDF(lines []string) ([]byte, error) {
	streamContent := buildPDFTextStream(lines)

	obj1 := "<< /Type /Catalog /Pages 2 0 R >>"
	obj2 := "<< /Type /Pages /Kids [3 0 R] /Count 1 >>"
	obj3 := "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>"
	obj4 := fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len([]byte(streamContent)), streamContent)
	obj5 := "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"

	objects := []string{obj1, obj2, obj3, obj4, obj5}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects)+1)

	for i, objectContent := range objects {
		objectID := i + 1
		offsets[objectID] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", objectID, objectContent)
	}

	xrefOffset := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(objects)+1)
	buf.WriteString("0000000000 65535 f \n")

	for i := 1; i <= len(objects); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}

	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefOffset)

	return buf.Bytes(), nil
}

func buildPDFTextStream(lines []string) string {
	var streamBuilder strings.Builder
	streamBuilder.WriteString("BT\n")
	streamBuilder.WriteString("/F1 12 Tf\n")
	streamBuilder.WriteString("50 790 Td\n")
	streamBuilder.WriteString("16 TL\n")

	for index, line := range lines {
		if index > 0 {
			streamBuilder.WriteString("T*\n")
		}
		streamBuilder.WriteString("(")
		streamBuilder.WriteString(escapePDFText(line))
		streamBuilder.WriteString(") Tj\n")
	}

	streamBuilder.WriteString("ET\n")
	return streamBuilder.String()
}

func escapePDFText(value string) string {
	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "(", "\\(")
	escaped = strings.ReplaceAll(escaped, ")", "\\)")
	return escaped
}

func parseAutomaticSignaturePermissions(body []byte) ([]automaticSignaturePermissionResponse, error) {
	var direct []automaticSignaturePermissionResponse
	if err := json.Unmarshal(body, &direct); err == nil {
		return direct, nil
	}

	var wrapped struct {
		Data        *[]automaticSignaturePermissionResponse `json:"data"`
		Permissions *[]automaticSignaturePermissionResponse `json:"permissions"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		if wrapped.Data != nil {
			return *wrapped.Data, nil
		}
		if wrapped.Permissions != nil {
			return *wrapped.Permissions, nil
		}
	}

	return nil, fmt.Errorf("unexpected response format from automatic-signature endpoint")
}

func parseAutomaticSignaturePermission(body []byte) (*automaticSignaturePermissionResponse, error) {
	var direct automaticSignaturePermissionResponse
	if err := json.Unmarshal(body, &direct); err == nil && direct.ID != "" {
		return &direct, nil
	}

	var wrapped struct {
		Data       *automaticSignaturePermissionResponse `json:"data"`
		Permission *automaticSignaturePermissionResponse `json:"permission"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		if wrapped.Data != nil {
			return wrapped.Data, nil
		}
		if wrapped.Permission != nil {
			return wrapped.Permission, nil
		}
	}

	return nil, fmt.Errorf("unexpected response format for automatic-signature permission")
}

func generateRandomPassword() (string, error) {
	buffer := make([]byte, 18)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	// 24 caracteres, alto entropia e válido para payload JSON.
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func defaultUserNameFromEmail(email string) string {
	localPart := strings.TrimSpace(email)
	if idx := strings.Index(localPart, "@"); idx > 0 {
		localPart = localPart[:idx]
	}

	localPart = strings.ReplaceAll(localPart, ".", " ")
	localPart = strings.ReplaceAll(localPart, "_", " ")
	localPart = strings.TrimSpace(localPart)
	if localPart == "" {
		return "Signatario Auto Signature"
	}

	return localPart
}

func hasVertcStatusCode(err error, statusCode int) bool {
	var vertErr *VertcAssinaturasError
	if !errors.As(err, &vertErr) {
		return false
	}

	return vertErr.StatusCode == statusCode
}

func isUserAlreadyExistsError(err error) bool {
	message := strings.ToLower(err.Error())

	if hasVertcStatusCode(err, http.StatusConflict) {
		return true
	}

	return strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "users_email_key") ||
		strings.Contains(message, "already exists") ||
		strings.Contains(message, "já existe")
}

func isRecipientUserNotFoundError(err error) bool {
	message := strings.ToLower(err.Error())

	if !hasVertcStatusCode(err, http.StatusNotFound) {
		return false
	}

	return strings.Contains(message, "destinat") ||
		strings.Contains(message, "recipient") ||
		strings.Contains(message, "não encontrado") ||
		strings.Contains(message, "nao encontrado") ||
		strings.Contains(message, "not found")
}

func isEnvelopeMissingDocumentsForSendError(err error) bool {
	message := strings.ToLower(err.Error())

	if !hasVertcStatusCode(err, http.StatusNotFound) {
		return false
	}

	return strings.Contains(message, "nenhum documento encontrado") ||
		strings.Contains(message, "nenhum documento") ||
		strings.Contains(message, "no document")
}

// IsAutomaticSignaturePermissionAlreadyExistsError verifica erro de permissão já existente.
func IsAutomaticSignaturePermissionAlreadyExistsError(err error) bool {
	message := strings.ToLower(err.Error())

	if !hasVertcStatusCode(err, http.StatusConflict) {
		return false
	}

	return strings.Contains(message, "assinatura automática ativa") ||
		strings.Contains(message, "automatic signature") ||
		strings.Contains(message, "already exists") ||
		strings.Contains(message, "já existe")
}
