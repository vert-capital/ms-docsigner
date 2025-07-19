# Sistema de Autenticação

Esta documentação detalha o sistema de autenticação JWT implementado no ms-docsigner, incluindo configuração, uso e troubleshooting.

## Visão Geral

O ms-docsigner utiliza **JWT (JSON Web Tokens)** para autenticação e autorização. Todos os endpoints da API, exceto os de health check, requerem um token JWT válido.

### Características do Sistema
- **Tipo**: Bearer Token Authentication
- **Algoritmo**: HMAC SHA-256 (HS256)
- **Localização**: Header `Authorization`
- **Formato**: `Bearer <token>`

---

## Headers Obrigatórios

### 1. Authorization Header
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

### 2. Content-Type Header (para requests com body)
```
Content-Type: application/json
```

### 3. Accept Header (recomendado)
```
Accept: application/json
```

---

## Estrutura do Token JWT

### Header
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

### Payload (Claims)
```json
{
  "sub": "1234567890",        // Subject (ID do usuário)
  "name": "John Doe",         // Nome do usuário
  "email": "user@example.com", // E-mail do usuário
  "role": "user",             // Papel do usuário
  "iat": 1516239022,          // Issued At (timestamp)
  "exp": 1516242622           // Expiration (timestamp)
}
```

### Signature
```
HMACSHA256(
  base64UrlEncode(header) + "." +
  base64UrlEncode(payload),
  secret
)
```

---

## Exemplo de Uso

### 1. Request Básico com Autenticação

```bash
curl -X GET https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -H "Accept: application/json"
```

### 2. Request com Body (POST)

```bash
curl -X POST https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "name": "Meu Documento",
    "file_content_base64": "JVBERi0xLjQKM..."
  }'
```

### 3. Usando Variáveis de Ambiente

```bash
# Definir token
export JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Usar em requests
curl -X GET https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json"
```

---

## Implementação em Diferentes Linguagens

### JavaScript (fetch)

```javascript
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';

fetch('https://api.ms-docsigner.com/api/v1/documents', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
})
.then(response => response.json())
.then(data => console.log(data))
.catch(error => console.error('Error:', error));
```

### JavaScript (axios)

```javascript
const axios = require('axios');

const api = axios.create({
  baseURL: 'https://api.ms-docsigner.com',
  headers: {
    'Authorization': `Bearer ${process.env.JWT_TOKEN}`,
    'Content-Type': 'application/json'
  }
});

// Fazer request
api.get('/api/v1/documents')
  .then(response => console.log(response.data))
  .catch(error => console.error(error.response.data));
```

### Python (requests)

```python
import requests
import os

token = os.getenv('JWT_TOKEN')
headers = {
    'Authorization': f'Bearer {token}',
    'Content-Type': 'application/json',
    'Accept': 'application/json'
}

response = requests.get(
    'https://api.ms-docsigner.com/api/v1/documents',
    headers=headers
)

if response.status_code == 200:
    data = response.json()
    print(data)
else:
    print(f'Error: {response.status_code} - {response.text}')
```

### Go

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
)

func main() {
    token := os.Getenv("JWT_TOKEN")
    
    client := &http.Client{}
    req, _ := http.NewRequest("GET", "https://api.ms-docsigner.com/api/v1/documents", nil)
    
    req.Header.Add("Authorization", "Bearer "+token)
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Accept", "application/json")
    
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}
```

### PHP

```php
<?php
$token = $_ENV['JWT_TOKEN'];

$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, 'https://api.ms-docsigner.com/api/v1/documents');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Authorization: Bearer ' . $token,
    'Content-Type: application/json',
    'Accept: application/json'
]);

$response = curl_exec($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

if ($httpCode == 200) {
    $data = json_decode($response, true);
    print_r($data);
} else {
    echo "Error: $httpCode - $response\n";
}
?>
```

---

## Middleware de Autenticação

### Como Funciona

1. **Extração do Token**: O middleware extrai o token do header `Authorization`
2. **Validação**: Verifica a assinatura e validade do token
3. **Decodificação**: Extrai as informações do usuário do payload
4. **Autorização**: Verifica se o usuário tem permissão para acessar o endpoint

### Fluxo de Validação

```
Request → Extract Token → Validate Signature → Check Expiration → Authorize → Process Request
    ↓            ↓               ↓                    ↓              ↓            ↓
   401          401             401                  401           403          200
```

---

## Códigos de Erro de Autenticação

### 401 Unauthorized

#### Token Ausente
```json
{
  "error": "Unauthorized",
  "message": "Authorization header is required"
}
```

#### Token Malformado
```json
{
  "error": "Unauthorized", 
  "message": "Invalid token format. Expected: Bearer <token>"
}
```

#### Token Inválido
```json
{
  "error": "Unauthorized",
  "message": "Invalid token signature"
}
```

#### Token Expirado
```json
{
  "error": "Unauthorized",
  "message": "Token has expired"
}
```

### 403 Forbidden

#### Permissões Insuficientes
```json
{
  "error": "Forbidden",
  "message": "Insufficient permissions to access this resource"
}
```

---

## Boas Práticas de Segurança

### 1. Armazenamento do Token

**✅ Recomendado:**
- Variáveis de ambiente
- Armazenamento seguro (Keychain, Vault)
- Sessões server-side

**❌ Evitar:**
- Hardcoding no código
- localStorage (apenas para testes)
- URLs ou logs

### 2. Transmissão

**✅ Sempre:**
- HTTPS/TLS em produção
- Headers seguros
- Validação de certificados

**❌ Nunca:**
- HTTP em produção
- Query parameters
- Logs de debug com tokens

### 3. Renovação de Tokens

```javascript
// Exemplo de interceptor para renovação automática
axios.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401) {
      // Token expirado, renovar automaticamente
      const newToken = await renewToken();
      error.config.headers['Authorization'] = `Bearer ${newToken}`;
      return axios.request(error.config);
    }
    return Promise.reject(error);
  }
);
```

---

## Troubleshooting Comum

### Problema: "Authorization header is required"

**Causa**: Header Authorization ausente
**Solução**:
```bash
# ❌ Incorreto
curl -X GET https://api.ms-docsigner.com/api/v1/documents

# ✅ Correto
curl -X GET https://api.ms-docsigner.com/api/v1/documents \
  -H "Authorization: Bearer your-token-here"
```

### Problema: "Invalid token format"

**Causa**: Formato do token incorreto
**Solução**:
```bash
# ❌ Incorreto
-H "Authorization: your-token-here"
-H "Authorization: JWT your-token-here"

# ✅ Correto
-H "Authorization: Bearer your-token-here"
```

### Problema: "Token has expired"

**Causa**: Token JWT expirado
**Solução**: Obter um novo token válido

### Problema: Timeout ou conexão negada

**Causa**: Possível problema de rede ou proxy
**Solução**:
```bash
# Verificar conectividade
curl -v https://api.ms-docsigner.com/health

# Verificar proxy (se aplicável)
curl --proxy-user user:pass --proxy proxy.company.com:8080 \
  https://api.ms-docsigner.com/api/v1/documents
```

---

## Configuração para Diferentes Ambientes

### Desenvolvimento
```bash
export JWT_TOKEN="dev-token-here"
export API_BASE_URL="https://api-dev.ms-docsigner.com"
```

### Staging
```bash
export JWT_TOKEN="staging-token-here" 
export API_BASE_URL="https://api-staging.ms-docsigner.com"
```

### Produção
```bash
export JWT_TOKEN="prod-token-here"
export API_BASE_URL="https://api.ms-docsigner.com"
```

---

## Validação de Token (para desenvolvedores)

### Ferramenta Online
Use [jwt.io](https://jwt.io) para decodificar e validar tokens durante desenvolvimento.

### Comando local
```bash
# Decodificar payload do token (base64)
echo "eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ" | base64 -d
```

### Script de validação
```bash
#!/bin/bash
TOKEN="$1"

if [ -z "$TOKEN" ]; then
  echo "Usage: $0 <jwt-token>"
  exit 1
fi

# Extrair payload (segunda parte do JWT)
PAYLOAD=$(echo $TOKEN | cut -d'.' -f2)

# Adicionar padding se necessário
case $((${#PAYLOAD} % 4)) in
  2) PAYLOAD="${PAYLOAD}==" ;;
  3) PAYLOAD="${PAYLOAD}=" ;;
esac

# Decodificar e formatar
echo $PAYLOAD | base64 -d 2>/dev/null | jq .
```

---

## Integração com Postman

### 1. Configurar Collection Variable

Em sua collection do Postman:
1. Vá em **Variables**
2. Adicione uma variável `jwt_token`
3. Cole seu token no campo **Current Value**

### 2. Configurar Authorization

Em cada request:
1. Vá na aba **Authorization**
2. Selecione **Type: Bearer Token**
3. No campo **Token**, use: `{{jwt_token}}`

### 3. Script de Pre-request (opcional)

Para renovação automática:
```javascript
// Pre-request Script
const token = pm.variables.get("jwt_token");

if (!token) {
    console.log("JWT token not found");
    return;
}

// Verificar se token está próximo do vencimento
const payload = JSON.parse(atob(token.split('.')[1]));
const exp = payload.exp * 1000; // Converter para ms
const now = Date.now();
const timeUntilExpiry = exp - now;

if (timeUntilExpiry < 300000) { // Menos de 5 minutos
    console.log("Token expiring soon, consider refreshing");
}
```

---

## Segurança Avançada

### Rate Limiting
A API implementa rate limiting baseado no token JWT. Evite fazer muitas requisições simultâneas.

### CORS
Para aplicações web, certifique-se de que seu domínio está na lista de origens permitidas.

### Headers de Segurança
A API retorna headers de segurança padrão:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`

---

## Referências

- [RFC 7519 - JSON Web Token (JWT)](https://tools.ietf.org/html/rfc7519)
- [RFC 6750 - The OAuth 2.0 Authorization Framework: Bearer Token Usage](https://tools.ietf.org/html/rfc6750)
- [OWASP JWT Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)