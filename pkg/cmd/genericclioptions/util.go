package genericclioptions

import (
	"fmt"
	"os"
	"strings"

	"github.com/openshift/odo/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Runnable interface {
	Complete(name string, cmd *cobra.Command, args []string) error
	Validate() error
	Run() error
}

func GenericRun(o Runnable, cmd *cobra.Command, args []string) {

	// CheckMachineReadableOutput
	// fixes / checks all related machine readable output functions
	// CheckMachineReadableOutputCommand(cmd)

	// Run completion, validation and run.
	LogErrorAndExit(o.Complete(cmd.Name(), cmd, args), "")
	LogErrorAndExit(o.Validate(), "")
	LogErrorAndExit(o.Run(), "")
}

// // CheckMachineReadableOutputCommand performs machine-readable output functions required to
// // have it work correctly
// func CheckMachineReadableOutputCommand(cmd *cobra.Command) {

// 	// Get the needed values
// 	outputFlag := pflag.Lookup("o")
// 	hasFlagChanged := outputFlag != nil && outputFlag.Changed
// 	machineOutput := cmd.Annotations["machineoutput"]

// 	// Check the valid output
// 	if hasFlagChanged && outputFlag.Value.String() != "json" {
// 		log.Error("Please input a valid output format for -o, available format: json")
// 		os.Exit(1)
// 	}

// 	// Check that if -o json has been passed, that the command actually USES json.. if not, error out.
// 	if hasFlagChanged && outputFlag.Value.String() == "json" && machineOutput == "" {

// 		// By default we "disable" logging, so undisable it so that the below error can be shown.
// 		_ = flag.Set("o", "")

// 		// Output the error
// 		log.Error("Machine readable output is not yet implemented for this command")
// 		os.Exit(1)
// 	}

// 	// Before running anything, we will make sure that no verbose output is made
// 	// This is a HACK to manually override `-v 4` to `-v 0` (in which we have no klog.V(0) in our code...
// 	// in order to have NO verbose output when combining both `-o json` and `-v 4` so json output
// 	// is not malformed / mixed in with normal logging
// 	if log.IsJSON() {
// 		_ = flag.Set("v", "0")
// 	} else {
// 		// Override the logging level by the value (if set) by the ODO_LOG_LEVEL env
// 		// The "-v" flag set on command line will take precedence over ODO_LOG_LEVEL env
// 		v := flag.CommandLine.Lookup("v").Value.String()
// 		if level, ok := os.LookupEnv("ODO_LOG_LEVEL"); ok && v == "0" {
// 			_ = flag.CommandLine.Set("v", level)
// 		}
// 	}
// }

// LogErrorAndExit prints the cause of the given error and exits the code with an
// exit code of 1.
// If the context is provided, then that is printed, if not, then the cause is
// detected using errors.Cause(err)
// *If* we are using the global json parameter, we instead output the json output
func LogErrorAndExit(err error, context string, a ...interface{}) {

	if err != nil {

		if context == "" {
			log.Error(errors.Cause(err))
		} else {
			printstring := fmt.Sprintf("%s%s", strings.Title(context), "\nError: %v")
			log.Errorf(printstring, err)
		}

		// Always exit 1 anyways
		os.Exit(1)

	}
}
