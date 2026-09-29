package get_gate_candidates_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetGateCandidates(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GetGateCandidates Usecase Suite")
}
