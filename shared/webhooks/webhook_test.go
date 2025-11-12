package webhooks

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSignature(t *testing.T) {
	service := &Service{}
	payload := []byte(`{"test": "data"}`)
	secret := "my-secret-key"

	signature1 := service.generateSignature(payload, secret)
	signature2 := service.generateSignature(payload, secret)

	// Same payload and secret should generate same signature
	assert.Equal(t, signature1, signature2)
	assert.NotEmpty(t, signature1)
	assert.Len(t, signature1, 64) // SHA256 hex is 64 chars
}

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"test": "data"}`)
	secret := "my-secret-key"

	signature := generateSignatureStatic(payload, secret)

	tests := []struct {
		name      string
		payload   []byte
		signature string
		secret    string
		expected  bool
	}{
		{
			name:      "valid signature",
			payload:   payload,
			signature: signature,
			secret:    secret,
			expected:  true,
		},
		{
			name:      "wrong secret",
			payload:   payload,
			signature: signature,
			secret:    "wrong-secret",
			expected:  false,
		},
		{
			name:      "wrong payload",
			payload:   []byte(`{"different": "data"}`),
			signature: signature,
			secret:    secret,
			expected:  false,
		},
		{
			name:      "wrong signature",
			payload:   payload,
			signature: "invalid-signature",
			secret:    secret,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifySignature(tt.payload, tt.signature, tt.secret)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScheduleRetry(t *testing.T) {
	service := &Service{
		maxRetries: 5,
	}

	tests := []struct {
		name            string
		attempt         int
		expectedStatus  string
		shouldHaveRetry bool
	}{
		{
			name:            "first attempt",
			attempt:         1,
			expectedStatus:  "pending",
			shouldHaveRetry: true,
		},
		{
			name:            "third attempt",
			attempt:         3,
			expectedStatus:  "pending",
			shouldHaveRetry: true,
		},
		{
			name:            "max attempts",
			attempt:         5,
			expectedStatus:  "failed",
			shouldHaveRetry: false,
		},
		{
			name:            "exceeded max",
			attempt:         6,
			expectedStatus:  "failed",
			shouldHaveRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := &Delivery{
				ID:      uuid.New(),
				Attempt: tt.attempt,
			}

			service.scheduleRetry(delivery)

			assert.Equal(t, tt.expectedStatus, delivery.Status)
			if tt.shouldHaveRetry {
				assert.NotNil(t, delivery.NextRetryAt)
			} else {
				assert.Nil(t, delivery.NextRetryAt)
			}
		})
	}
}
