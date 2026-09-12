package notify_overdue_sites_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNotifyOverdueSites(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NotifyOverdueSites Usecase Suite")
}
