package poll_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPoll(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Poll Suite")
}
