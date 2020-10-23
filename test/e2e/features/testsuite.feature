Feature: Testsuite test
Quentin check whether their testsuite works properly.

  Scenario: Contains
     When executing "go help" succeeds
     Then stdout should contain
     """
     Go is a tool for managing Go source code.
     """

  Scenario: Not Contains
     When executing "go help" succeeds
     Then stdout should not contain "Error"

  Scenario: Equals
     When executing "go help" succeeds
     Then exitcode should equal "0"

  Scenario: Not Equals
     When executing "go notexist" fails
     Then exitcode should not equal "0"

  Scenario: Matches
     When executing "go version" succeeds
     Then stdout should match "go1\.\d+\.\d+"
