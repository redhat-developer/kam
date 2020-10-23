package helper

import (
	"fmt"
	"strings"
)

var scenarioVariables []scenarioVariable

type scenarioVariable struct {
	Name  string
	Value string
}

func ProcessScenarioVariables(command string) string {
	for _, variable := range scenarioVariables {
		command = strings.Replace(command, fmt.Sprintf("$(%s)", variable.Name), variable.Value, -1)
	}

	return command
}

func ClearScenarioVariables() {
	scenarioVariables = nil
}

func SetScenarioVariable(name string, value string) {
	newVariable := scenarioVariable{
		name,
		value,
	}

	scenarioVariables = append([]scenarioVariable{newVariable}, scenarioVariables...)
}
