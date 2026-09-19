package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGoogleLogin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Google Login Usecase Suite")
}
