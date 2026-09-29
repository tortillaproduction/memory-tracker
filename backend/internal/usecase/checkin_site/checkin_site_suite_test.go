package checkin_site_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCheckinSite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CheckinSite Usecase Suite")
}
