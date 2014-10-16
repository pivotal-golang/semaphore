package semaphore_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-golang/semaphore"
	"github.com/tedsuo/ifrit"
)

var _ = Describe("Semaphore", func() {
	var semaphore Semaphore
	var semaphoreProcess ifrit.Process

	BeforeEach(func() {
		semaphore = New(1, 2)
		semaphoreProcess = ifrit.Invoke(semaphore)
	})

	AfterEach(func() {
		semaphoreProcess.Signal(os.Kill)
	})

	Context("when maxInFlight has not yet been reached", func() {
		It("does not block when acquiring once", func(done Done) {
			defer close(done)
			_, err := semaphore.Acquire()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("when maxInFlight is reached", func() {
		var resource Resource

		BeforeEach(func() {
			var err error
			resource, err = semaphore.Acquire()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("blocks when trying to acquire", func(done Done) {
			defer close(done)

			acquired := make(chan struct{})
			semaphore := semaphore
			go func() {
				semaphore.Acquire()
				close(acquired)
			}()

			Consistently(acquired).ShouldNot(BeClosed())
		})

		It("processes requests in the order that they are received", func(done Done) {
			defer close(done)

			var orderOfProcessingPendingRequests []int
			request1 := make(chan struct{})
			request2 := make(chan struct{})

			go func() {
				resource1, err := semaphore.Acquire()
				Expect(err).ShouldNot(HaveOccurred())
				orderOfProcessingPendingRequests = append(orderOfProcessingPendingRequests, 1)
				resource1.Release()
				close(request1)
			}()

			go func() {
				resource2, err := semaphore.Acquire()
				Expect(err).ShouldNot(HaveOccurred())
				orderOfProcessingPendingRequests = append(orderOfProcessingPendingRequests, 2)
				resource2.Release()
				close(request2)
			}()

			resource.Release()

			<-request1
			<-request2

			Expect(orderOfProcessingPendingRequests).To(Equal([]int{1, 2}))
		})

		Context("and a request completes", func() {
			BeforeEach(func() {
				resource.Release()
			})

			It("does not block when acquiring, releasing and acquiring again", func(done Done) {
				defer close(done)
				semaphore.Acquire()
			})
		})

		Context("when maxPanding is reached", func() {
			BeforeEach(func() {
				go semaphore.Acquire()
				go semaphore.Acquire()
			})

			It("returns an error trying to acquire", func(done Done) {
				defer close(done)

				_, err := semaphore.Acquire()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Cannot queue request, maxPending reached: 2"))
			})
		})
	})
})
