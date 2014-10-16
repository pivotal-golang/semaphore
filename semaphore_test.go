package semaphore_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-golang/semaphore"
)

var _ = Describe("Semaphore", func() {
	var semaphore Semaphore

	BeforeEach(func() {
		semaphore = New(1, 2)
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

		Context("when maxPending is reached", func() {
			It("returns an error trying to acquire", func(done Done) {
				defer close(done)

				errChan := make(chan error)
				for i := 0; i < 4; i++ {
					semaphore := semaphore
					go func() {
						_, err := semaphore.Acquire()
						errChan <- err
					}()
				}

				var err error
				Eventually(errChan).Should(Receive(&err))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Cannot queue request, maxPending reached: 2"))
			})
		})
	})

	Context("when trying to release twice", func() {
		It("returns an error", func(done Done) {
			defer close(done)

			resource, err := semaphore.Acquire()
			Expect(err).ToNot(HaveOccurred())

			err = resource.Release()
			Expect(err).ToNot(HaveOccurred())

			err = resource.Release()
			Expect(err).To(Equal(errors.New("Resource has already been released")))
		})
	})
})
