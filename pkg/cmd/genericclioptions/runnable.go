package genericclioptions

import (
	"fmt"
	"os"
	"strings"

	"github.com/openshift/odo/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Runnable interface represents a command
type Runnable interface {
	Complete(name string, cmd *cobra.Command, args []string) error
	Validate() error
	Run() error
}

// GenericRun executes the Runnable methods in the right order
func GenericRun(o Runnable, cmd *cobra.Command, args []string) {
	// Run completion, validation and run.
	logErrorAndExit(o.Complete(cmd.Name(), cmd, args), "")
	logErrorAndExit(o.Validate(), "")
	logErrorAndExit(o.Run(), "")
}

// LogErrorAndExit prints the cause of the given error and exits the code with an
// exit code of 1.
// If the context is provided, then that is printed, if not, then the cause is
// detected using errors.Cause(err)
func logErrorAndExit(err error, context string, a ...interface{}) {
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
