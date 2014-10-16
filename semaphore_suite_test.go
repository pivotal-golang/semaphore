package semaphore

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSemaphore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Semaphore Suite")
}
