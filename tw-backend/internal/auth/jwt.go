package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims extends jwt.RegisteredClaims with custom fields.
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// TokenManager handles JWT generation and validation.
type TokenManager struct {
	signingKey    []byte
	encryptionKey []byte
}

// NewTokenManager creates a new TokenManager.
// signingKey must be at least 32 bytes for HS256 (recommended).
// encryptionKey must be exactly 32 bytes for AES-256.
func NewTokenManager(signingKey, encryptionKey []byte) (*TokenManager, error) {
	if len(encryptionKey) != 32 {
		return nil, errors.New("encryption key must be 32 bytes for AES-256")
	}
	return &TokenManager{
		signingKey:    signingKey,
		encryptionKey: encryptionKey,
	}, nil
}

// GenerateToken creates a signed and encrypted JWT.
func (tm *TokenManager) GenerateToken(userID, username string, roles []string) (string, error) {
	// 1. Prepare sensitive data
	sensitiveData := struct {
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	}{
		Username: username,
		Roles:    roles,
	}

	jsonData, err := json.Marshal(sensitiveData)
	if err != nil {
		return "", err
	}

	// 2. Encrypt sensitive data
	encryptedData, err := tm.encrypt(jsonData)
	if err != nil {
		return "", err
	}

	// 3. Create JWT claims
	claims := jwt.MapClaims{
		"sub":      userID,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
		"enc_data": base64.StdEncoding.EncodeToString(encryptedData),
	}

	// 4. Sign token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.signingKey)
}

// ValidateToken parses, decrypts, and validates a JWT.
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	// 1. Parse and validate signature
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims structure")
	}

	// 2. Extract and decrypt data
	encDataStr, ok := mapClaims["enc_data"].(string)
	if !ok {
		return nil, errors.New("missing encrypted data")
	}

	encData, err := base64.StdEncoding.DecodeString(encDataStr)
	if err != nil {
		return nil, errors.New("invalid base64 data")
	}

	decryptedJson, err := tm.decrypt(encData)
	if err != nil {
		return nil, errors.New("failed to decrypt data")
	}

	var sensitiveData struct {
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	}
	if err := json.Unmarshal(decryptedJson, &sensitiveData); err != nil {
		return nil, errors.New("failed to unmarshal decrypted data")
	}

	// 3. Construct full Claims object with safe type assertions
	sub, ok := mapClaims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid subject claim")
	}
	exp, ok := mapClaims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid expiration claim")
	}
	iat, ok := mapClaims["iat"].(float64)
	if !ok {
		return nil, errors.New("invalid issued-at claim")
	}

	claims := &Claims{
		UserID:   sub,
		Username: sensitiveData.Username,
		Roles:    sensitiveData.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "mud-platform",
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(time.Unix(int64(exp), 0)),
			IssuedAt:  jwt.NewNumericDate(time.Unix(int64(iat), 0)),
		},
	}

	return claims, nil
}

func (tm *TokenManager) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(tm.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (tm *TokenManager) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(tm.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
