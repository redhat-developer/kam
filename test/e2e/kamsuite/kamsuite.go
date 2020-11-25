package kamsuite

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// KAM related steps

	s.BeforeSuite(func() {
		fmt.Println("Before suite")
	})

	s.AfterSuite(func() {
		fmt.Println("After suite")
	})

	s.BeforeFeature(func(this *messages.GherkinDocument) {
		fmt.Println("Before feature")
	})

	s.AfterFeature(func(this *messages.GherkinDocument) {
		fmt.Println("After feature")
	})
}
