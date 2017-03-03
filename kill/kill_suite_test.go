package kill_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"testing"

	"github.com/onsi/gomega/gexec"
)

func TestKill(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kill Suite")
}

var easyPath string
var hardPath string

type testData struct {
	Easy string
	Hard string
}

var _ = SynchronizedBeforeSuite(func() []byte {
	easyPath, err := gexec.Build("github.com/cloudfoundry/sipid/kill/fixtures/easy_kill")
	Expect(err).NotTo(HaveOccurred())

	hardPath, err := gexec.Build("github.com/cloudfoundry/sipid/kill/fixtures/hard_kill")
	Expect(err).NotTo(HaveOccurred())

	bs, err := json.Marshal(testData{
		Easy: easyPath,
		Hard: hardPath,
	})
	Expect(err).NotTo(HaveOccurred())

	return []byte(bs)
}, func(data []byte) {
	var td testData

	err := json.Unmarshal(data, &td)
	Expect(err).NotTo(HaveOccurred())

	easyPath = td.Easy
	hardPath = td.Hard
})
