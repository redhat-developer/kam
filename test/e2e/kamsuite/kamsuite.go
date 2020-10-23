package kamsuite

import (
	"flag"
	"fmt"
	"os"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
	"github.com/redhat-developer/kam/test/e2e/helper"
)

var (
	testDir         string
	testRunDir      string
	testResultsDir  string
	testDefaultHome string
	testWithShell   string

	GodogFormat              string
	GodogTags                string
	GodogShowStepDefinitions bool
	GodogStopOnFailure       bool
	GodogNoColors            bool
	GodogPaths               string
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {
	// Executing commands
	s.Step(`^executing "(.*)"$`,
		ExecuteCommand)
	s.Step(`^executing "(.*)" (succeeds|fails)$`,
		ExecuteCommandSucceedsOrFails)

	// Command output verification
	s.Step(`^(stdout|stderr|exitcode) (?:should contain|contains) "(.*)"$`,
		CommandReturnShouldContain)
	s.Step(`^(stdout|stderr|exitcode) (?:should contain|contains)$`,
		CommandReturnShouldContainContent)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not contain "(.*)"$`,
		CommandReturnShouldNotContain)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does not) contain$`,
		CommandReturnShouldNotContainContent)

	s.Step(`^(stdout|stderr|exitcode) (?:should equal|equals) "(.*)"$`,
		CommandReturnShouldEqual)
	s.Step(`^(stdout|stderr|exitcode) (?:should equal|equals)$`,
		CommandReturnShouldEqualContent)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not equal "(.*)"$`,
		CommandReturnShouldNotEqual)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not equal$`,
		CommandReturnShouldNotEqualContent)

	s.Step(`^(stdout|stderr|exitcode) (?:should match|matches) "(.*)"$`,
		CommandReturnShouldMatch)
	s.Step(`^(stdout|stderr|exitcode) (?:should match|matches)`,
		CommandReturnShouldMatchContent)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not match "(.*)"$`,
		CommandReturnShouldNotMatch)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not match`,
		CommandReturnShouldNotMatchContent)

	s.Step(`^(stdout|stderr|exitcode) (?:should be|is) empty$`,
		CommandReturnShouldBeEmpty)
	s.Step(`^(stdout|stderr|exitcode) (?:should not be|is not) empty$`,
		CommandReturnShouldNotBeEmpty)

	s.Step(`^(stdout|stderr|exitcode) (?:should be|is) valid "([^"]*)"$`,
		ShouldBeInValidFormat)

	// Command output and execution: extra steps
	s.Step(`^with up to "(\d*)" retries with wait period of "(\d*(?:ms|s|m))" command "(.*)" output (should contain|contains|should not contain|does not contain) "(.*)"$`,
		ExecuteCommandWithRetry)
	s.Step(`^evaluating stdout of the previous command succeeds$`,
		ExecuteStdoutLineByLine)

	// Scenario variables
	// allows to set a scenario variable to the output values of minishift and oc commands
	// and then refer to it by $(NAME_OF_VARIABLE) directly in the text of feature file
	s.Step(`^setting scenario variable "(.*)" to the stdout from executing "(.*)"$`,
		SetScenarioVariableExecutingCommand)

	// Filesystem operations
	s.Step(`^creating directory "([^"]*)" succeeds$`,
		CreateDirectory)
	s.Step(`^creating file "([^"]*)" succeeds$`,
		CreateFile)
	s.Step(`^deleting directory "([^"]*)" succeeds$`,
		DeleteDirectory)
	s.Step(`^deleting file "([^"]*)" succeeds$`,
		DeleteFile)
	s.Step(`^directory "([^"]*)" should not exist$`,
		DirectoryShouldNotExist)
	s.Step(`^file "([^"]*)" should not exist$`,
		FileShouldNotExist)
	s.Step(`^file "([^"]*)" exists$`,
		FileExist)
	s.Step(`^file from "(.*)" is downloaded into location "(.*)"$`,
		DownloadFileIntoLocation)
	s.Step(`^writing text "([^"]*)" to file "([^"]*)" succeeds$`,
		WriteToFile)

	// File content checks
	s.Step(`^content of file "([^"]*)" should contain "([^"]*)"$`,
		FileContentShouldContain)
	s.Step(`^content of file "([^"]*)" should not contain "([^"]*)"$`,
		FileContentShouldNotContain)
	s.Step(`^content of file "([^"]*)" should equal "([^"]*)"$`,
		FileContentShouldEqual)
	s.Step(`^content of file "([^"]*)" should not equal "([^"]*)"$`,
		FileContentShouldNotEqual)
	s.Step(`^content of file "([^"]*)" should match "([^"]*)"$`,
		FileContentShouldMatchRegex)
	s.Step(`^content of file "([^"]*)" should not match "([^"]*)"$`,
		FileContentShouldNotMatchRegex)
	s.Step(`^content of file "([^"]*)" (?:should be|is) valid "([^"]*)"$`,
		FileContentIsInValidFormat)

	// Config file content, JSON and YAML
	s.Step(`"(JSON|YAML)" config file "(.*)" (contains|does not contain) key "(.*)" with value matching "(.*)"$`,
		ConfigFileContainsKeyMatchingValue)
	s.Step(`"(JSON|YAML)" config file "(.*)" (contains|does not contain) key "(.*)"$`,
		ConfigFileContainsKey)

	s.BeforeSuite(func() {
		err := PrepareForIntegrationTest()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})

	s.BeforeFeature(func(this *messages.GherkinDocument) {
		helper.LogMessage("info", fmt.Sprintf("----- Feature: %s -----", this.String()))
		StartHostShellInstance(testWithShell)
		helper.ClearScenarioVariables()
		err := CleanTestRunDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})

	s.BeforeScenario(func(this *messages.Pickle) {
		helper.LogMessage("info", fmt.Sprintf("----- Scenario: %s -----", this.Name))
		helper.LogMessage("info", fmt.Sprintf("----- Scenario Outline: %s -----", this.String()))
	})

	s.BeforeStep(func(this *messages.Pickle_PickleStep) {
		this.Text = helper.ProcessScenarioVariables(this.Text)
	})

	s.AfterScenario(func(*messages.Pickle, error) {
	})

	s.AfterFeature(func(*messages.GherkinDocument) {
		helper.LogMessage("info", "----- Cleaning after feature -----")
		CloseHostShellInstance()
	})

	s.AfterSuite(func() {
		helper.LogMessage("info", "----- Cleaning Up -----")
		err := helper.CloseLog()
		if err != nil {
			fmt.Println("Error closing the log:", err)
		}
	})
}

func ParseFlags() {
	flag.StringVar(&testDir, "test-dir", "out", "Path to the directory in which to execute the tests")
	flag.StringVar(&testWithShell, "test-shell", "", "Specifies shell to be used for the testing.")

	flag.StringVar(&GodogFormat, "godog.format", "pretty", "Sets which format godog will use")
	flag.StringVar(&GodogTags, "godog.tags", "", "Tags for godog test")
	flag.BoolVar(&GodogShowStepDefinitions, "godog.definitions", false, "")
	flag.BoolVar(&GodogStopOnFailure, "godog.stop-on-failure ", false, "Stop when failure is found")
	flag.BoolVar(&GodogNoColors, "godog.no-colors", false, "Disable colors in godog output")
	flag.StringVar(&GodogPaths, "godog.paths", "./features", "")

	flag.Parse()
}
