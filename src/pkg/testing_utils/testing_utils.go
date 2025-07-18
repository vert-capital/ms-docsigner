package testing_utils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Contains[T string | int](elems []T, item T) bool {
	for _, v := range elems {
		if v == item {
			return true
		}
	}
	return false
}

type TestRequest struct {
	Url        string
	Body       []byte
	Method     string
	Request    *http.Request
	Response   *httptest.ResponseRecorder
	HttpHandle gin.HandlerFunc
	T          *testing.T
	executed   bool
}

func (r *TestRequest) SetJsonBody(body any) {
	r.Body, _ = json.Marshal(body)
}

func (r *TestRequest) SetTextBody(body string) {
	r.Body = []byte(body)
}

func (r *TestRequest) Init() {
	r.executed = false
	r.Response = httptest.NewRecorder()

	r.Request, _ = http.NewRequest(r.Method, r.Url, nil)

	r.Request, _ = http.NewRequest(r.Method, r.Url, bytes.NewBuffer(r.Body))
}

func (r *TestRequest) SetHandle(httpHandle gin.HandlerFunc) {
	r.HttpHandle = httpHandle
}

func (r *TestRequest) Execute() {

	r.Init()

	r.Request.Header.Set("Content-Type", "application/json")

	gin.SetMode(gin.TestMode)
	router := gin.New()

	switch r.Method {
	case "GET":
		router.GET(r.Url, r.HttpHandle)
	case "POST":
		router.POST(r.Url, r.HttpHandle)
	case "PUT":
		router.PUT(r.Url, r.HttpHandle)
	case "DELETE":
		router.DELETE(r.Url, r.HttpHandle)
	}

	router.ServeHTTP(r.Response, r.Request)

	r.executed = true
}

func (r *TestRequest) AssertStatus(status int) {
	if !r.executed {
		r.Execute()
	}
	assert.Equal(r.T, status, r.Response.Code, "OK response is expected")
}

func (r *TestRequest) GetBody() []byte {
	if !r.executed {
		r.Execute()
	}
	return r.Response.Body.Bytes()
}

func (r *TestRequest) GetJsonResponse() map[string]any {
	if !r.executed {
		r.Execute()
	}
	var jsonResponse map[string]any
	json.Unmarshal(r.Response.Body.Bytes(), &jsonResponse)
	return jsonResponse
}

func (r *TestRequest) SetAuthorizationToken(firstName string, lastName string, email string) {
	r.Request.Header.Set("Authorization", "Bearer "+r.getUserToken(firstName, lastName, email))
	r.Request.WithContext(context.WithValue(r.Request.Context(), "user", "1"))
}

func (r *TestRequest) getUserToken(firstName string, lastName string, email string) string {
	// token, _, _ := entity.JWTTokenGenerator(
	// 	"1",
	// 	firstName,
	// 	lastName,
	// 	email,
	// )

	return "token"
}
