package middleware

import (
	"strings"
	"testing"
)

func TestMaskSensitiveFields(t *testing.T) {
	raw := []byte(`{"username":"u1","password":"abc","token":"t1","cluster":{"kubeconfig":"kcfg"}}`)
	masked := maskSensitiveFields(raw)

	if strings.Contains(masked, `"password":"abc"`) {
		t.Fatalf("password should be masked, got %s", masked)
	}
	if strings.Contains(masked, `"token":"t1"`) {
		t.Fatalf("token should be masked, got %s", masked)
	}
	if strings.Contains(masked, `"kubeconfig":"kcfg"`) {
		t.Fatalf("kubeconfig should be masked, got %s", masked)
	}
}
