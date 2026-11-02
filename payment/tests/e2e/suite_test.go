//go:build e2e

package e2e

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"inventory-go/platform/pkg/logger"
)

var (
	env *TestEnvironment

	suiteCtx    context.Context
	suiteCancel context.CancelFunc
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Payment Service E2E Test Suite")
}

var _ = BeforeSuite(func() {
	Expect(logger.Init(loggerLevelValue, true)).To(Succeed(), "expected the logger to be initialized")

	suiteCtx, suiteCancel = context.WithTimeout(context.Background(), testsTimeout)

	logger.Info(suiteCtx, "Starting the test environment...")

	var err error
	env, err = setupTestEnvironment(suiteCtx)
	Expect(err).ToNot(HaveOccurred(), "expected the test environment to start")
})

var _ = AfterSuite(func() {
	logger.Info(context.Background(), "Finishing the test suite")

	if env != nil {
		teardownTestEnvironment(suiteCtx, env)
	}

	if suiteCancel != nil {
		suiteCancel()
	}
})
