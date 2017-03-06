package poll_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/cloudfoundry/sipid/poll"
	"context"
	"time"
	"fmt"
)

var _ = Describe("Poll", func() {
	var (
		server *ghttp.Server
		serverURL string
	)

	pollingFrequency := 200 * time.Millisecond

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL = server.URL()
	})

	AfterEach(func() {
		server.Close()
	})

	failure := ghttp.RespondWith(500, nil)
	success := ghttp.RespondWith(200, nil)

	Context("when the endpoint is eventually successful", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				failure,
				failure,
				success,
			)
		})

		It("returns with no error", func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := poll.Poll(ctx, serverURL, pollingFrequency)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("timeouts", func() {
		Context("endpoint never reachable", func() {
			BeforeEach(func() {
				server.Close()
			})

			It("returns an error", func() {
				ctx, cancel := context.WithTimeout(context.Background(), 500 * time.Millisecond)
				defer cancel()

				err := poll.Poll(ctx, serverURL, pollingFrequency)
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("healthcheck unhealthy (url '%s' was not healthy after 2 attempts)", serverURL))))
			})
		})

		Context("endpoint only returns unsuccessful", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					failure,
					failure,
					failure,
				)
			})

			It("returns an error", func() {
				ctx, cancel := context.WithTimeout(context.Background(), 500 * time.Millisecond)
				defer cancel()

				err := poll.Poll(ctx, serverURL, pollingFrequency)
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("healthcheck unhealthy (url '%s' was not healthy after 2 attempts)", serverURL))))
			})
		})
	})

	Describe("invalid URL", func() {
		It("returns an error", func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := poll.Poll(ctx, ":health.example.com/health", pollingFrequency)
			Expect(err).To(MatchError(ContainSubstring("missing protocol scheme")))
		})
	})
})
