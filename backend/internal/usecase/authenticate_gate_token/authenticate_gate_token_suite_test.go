package authenticate_gate_token_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAuthenticateGateToken(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AuthenticateGateToken Usecase Suite")
}
