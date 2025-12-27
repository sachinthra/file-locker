package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterAndLogin(t *testing.T) {
	// 1. Test Register
	regPayload := `{"username":"testuser1","password":"password123","email":"test@test.com"}`
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(regPayload))
	w := httptest.NewRecorder()

	testAuth.HandleRegister(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp AuthResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.UserID)

	// 2. Test Login
	loginPayload := `{"username":"testuser1","password":"password123"}`
	reqLogin := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(loginPayload))
	wLogin := httptest.NewRecorder()

	testAuth.HandleLogin(wLogin, reqLogin)

	assert.Equal(t, http.StatusOK, wLogin.Code)
}

func TestLoginInvalidCredentials(t *testing.T) {
	loginPayload := `{"username":"wronguser","password":"wrongpassword"}`
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(loginPayload))
	w := httptest.NewRecorder()

	testAuth.HandleLogin(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
