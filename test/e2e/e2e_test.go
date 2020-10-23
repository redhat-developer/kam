package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/redhat-developer/kam/test/e2e/kamsuite"
)

func TestMain(m *testing.M) {
	parseFlags()

	status := godog.RunWithOptions("kam", func(s *godog.Suite) {
		getFeatureContext(s)
	}, godog.Options{
		Format:              kamsuite.GodogFormat,
		Paths:               strings.Split(kamsuite.GodogPaths, ","),
		Tags:                kamsuite.GodogTags,
		ShowStepDefinitions: kamsuite.GodogShowStepDefinitions,
		StopOnFailure:       kamsuite.GodogStopOnFailure,
		NoColors:            kamsuite.GodogNoColors,
	})

	os.Exit(status)
}

func getFeatureContext(s *godog.Suite) {
	kamsuite.FeatureContext(s)
}

func parseFlags() {
	kamsuite.ParseFlags()
}
