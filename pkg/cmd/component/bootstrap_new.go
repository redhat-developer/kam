package bootstrapnew

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/openshift/odo/pkg/log"
	"github.com/redhat-developer/kam/pkg/cmd/component/cmd/ui"
	"github.com/redhat-developer/kam/pkg/cmd/genericclioptions"
	"github.com/redhat-developer/kam/pkg/cmd/utility"
	"github.com/redhat-developer/kam/pkg/pipelines/accesstoken"
	pipelines "github.com/redhat-developer/kam/pkg/pipelines/component"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	ComponentRecommendedCommandName = "bootstrap-new"
	componentNameFlag               = "component-name"
	appliactionNameFlag             = "application-name"
	gitRepoURLFlag                  = "git-repo-url"
	secretFlag                      = "secret"
)

var (
	bootstrapExampleC = ktemplates.Examples(`
    # Bootstrap-New OpenShift pipelines.
		kam bootstrap-new --git-repo-url https://github.com/<your organization>/gitops.git --application-name <name of application> --component-name <name of component> --secret <your git access token> --output <path to write GitOps resources> --push-to-git=true
		
    %[1]s 
    `)

	bootstrapLongDescC  = ktemplates.LongDesc(`New Bootstrap Command`)
	bootstrapShortDescC = `New Bootstrap Command Application Configuration`
)

// BootstrapNewParameters encapsulates the parameters for the kam pipelines init command.
type BootstrapNewParameters struct {
	*pipelines.BootstrapNewOptions
	Interactive bool
}

type drivers []string

var (
	supportedDrivers = drivers{
		"github",
		"gitlab",
	}
)

func (d drivers) supported(s string) bool {
	for _, v := range d {
		if s == v {
			return true
		}
	}
	return false
}

// NewBootstrapNewParameters bootsraps a Bootstrap Parameters instance.
func NewBootstrapNewParameters() *BootstrapNewParameters {
	return &BootstrapNewParameters{
		BootstrapNewOptions: &pipelines.BootstrapNewOptions{},
	}
}

// Complete completes BootstrapNewParameters after they've been created.
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *BootstrapNewParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	_, err := utility.NewClient()
	if err != nil {
		return err
	}

	if io.PrivateRepoURLDriver != "" {
		host, err := accesstoken.HostFromURL(io.GitRepoURL)
		if err != nil {
			return err
		}
		identifier := factory.NewDriverIdentifier(factory.Mapping(host, io.PrivateRepoURLDriver))
		factory.DefaultIdentifier = identifier
	}

	if cmd.Flags().NFlag() == 0 || io.Interactive {
		return initiateInteractiveModeForBootstrapNewCommand(io, cmd)
	}
	addGitURLSuffixIfNecessary(io)
	return nonInteractiveModeBootstrapNew(io)
}
func addGitURLSuffixIfNecessary(io *BootstrapNewParameters) {
	io.GitRepoURL = utility.AddGitSuffixIfNecessary(io.GitRepoURL)
}

// nonInteractiveMode gets triggered if a flag is passed, checks for mandatory flags.
func nonInteractiveModeBootstrapNew(io *BootstrapNewParameters) error {
	mandatoryFlags := map[string]string{componentNameFlag: io.ComponentName, appliactionNameFlag: io.ApplicationName, gitRepoURLFlag: io.GitRepoURL, secretFlag: io.Secret}
	if err := checkMandatoryFlags(mandatoryFlags); err != nil {
		return err
	}
	err := ui.ValidateTargetPort(io.TargetPort)
	if err != nil {
		return fmt.Errorf("%v Target Port is not valid", io.TargetPort)
	}
	if io.PrivateRepoURLDriver != "" {
		if !supportedDrivers.supported(io.PrivateRepoURLDriver) {
			return fmt.Errorf("invalid driver type: %q", io.PrivateRepoURLDriver)
		}
	}
	err = setAccessToken(io)
	if err != nil {
		return err
	}

	return nil
}

// Checking the mandatory flags & reusing missingFlagErr in Bootstrap.go
func checkMandatoryFlags(flags map[string]string) error {
	missingFlags := []string{}
	mandatoryFlags := []string{componentNameFlag, appliactionNameFlag, gitRepoURLFlag, secretFlag}
	for _, flag := range mandatoryFlags {
		if flags[flag] == "" {
			missingFlags = append(missingFlags, fmt.Sprintf("%q", flag))
		}
	}
	if len(missingFlags) > 0 {
		return missingFlagErr(missingFlags)
	}
	return nil
}

func missingFlagErr(flags []string) error {
	return fmt.Errorf("required flag(s) %s not set", strings.Join(flags, ", "))
}

//Interactive mode for Bootstrap-mew Command
func initiateInteractiveModeForBootstrapNewCommand(io *BootstrapNewParameters, cmd *cobra.Command) error {
	log.Progressf("\nStarting interactive prompt\n")
	//Checks for mandatory flags
	promp := !ui.UseDefaultValuesComponent()
	if io.ApplicationName == "" {
		io.ApplicationName = ui.AddApplicationName()
	}
	if io.ComponentName == "" {
		io.ComponentName = ui.AddComponentName()
	}
	if io.GitRepoURL == "" {
		io.GitRepoURL = ui.EnterGitRepoURL()
	}
	io.GitRepoURL = utility.AddGitSuffixIfNecessary(io.GitRepoURL)
	if !isKnownDriverURL(io.GitRepoURL) {
		io.PrivateRepoURLDriver = ui.SelectPrivateRepoDriver()
		host, err := accesstoken.HostFromURL(io.GitRepoURL)
		if err != nil {
			return fmt.Errorf("failed to parse the gitops url: %w", err)
		}
		identifier := factory.NewDriverIdentifier(factory.Mapping(host, io.PrivateRepoURLDriver))
		factory.DefaultIdentifier = identifier
	}
	// We are checking if any existing token is present.
	//If not we ask the uer to pass the token.
	//EnterGitSecret is just validating length for now.
	secret, err := accesstoken.GetAccessToken(io.GitRepoURL)
	if err != nil && err != keyring.ErrNotFound {
		return err
	}
	if secret == "" { // We must prompt for the token
		if io.Secret == "" {
			io.Secret = ui.EnterGitSecret(io.GitRepoURL)
		}
		if !cmd.Flag("save-token-keyring").Changed {
			io.SaveTokenKeyRing = ui.UseKeyringRingSvc()
		}
		setAccessToken(io)
	} else {
		io.Secret = secret
	}

	//Optional flags
	if promp {
		io.TargetPort = ui.AddTargetPort()
	}

	if !cmd.Flag("push-to-git").Changed && promp {
		io.PushToGit = ui.SelectOptionPushToGit()
	}

	outputPathOverridden := cmd.Flag("output").Changed
	if !outputPathOverridden {
		// Override the default path to be ./{gitops repo name}
		repoName, err := repoFromURL(io.GitRepoURL)
		if err != nil {
			repoName = "gitops"
		}
		io.Output = filepath.Join(".", repoName)
	}

	appFs := ioutils.NewFilesystem()
	io.Output, io.Overwrite = ui.VerifyOutput(appFs, io.Output, io.Overwrite, io.ApplicationName, outputPathOverridden, promp)
	if !io.Overwrite {
		if ui.PathExists(appFs, filepath.Join(io.Output, io.ApplicationName)) {
			return fmt.Errorf("the secrets folder located as a sibling of the output folder %s already exists. Delete or rename the secrets folder and try again", io.Output)
		}
		if io.PushToGit && ui.PathExists(appFs, filepath.Join(io.Output, ".git")) {
			return fmt.Errorf("the .git folder in output path %s already exists. Delete or rename the .git folder and try again", io.Output)
		}
	}

	return nil
}
func repoFromURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Path, "/")
	return strings.TrimSuffix(parts[len(parts)-1], ".git"), nil
}

//
func setAccessToken(io *BootstrapNewParameters) error {
	if io.SaveTokenKeyRing {
		err := accesstoken.SetAccessToken(io.GitRepoURL, io.Secret)
		if err != nil {
			return err
		}
	}
	if io.Secret == "" {
		secret, err := accesstoken.GetAccessToken(io.GitRepoURL)
		if err != nil {
			return fmt.Errorf("unable to use access-token from keyring/env-var: %v, please pass a valid token to --git-host-access-token", err)
		}
		io.Secret = secret
	}
	return nil
}

// Validate validates the parameters of the CompomemtParameters.
func (io *BootstrapNewParameters) Validate() error {

	gr, err := url.Parse(io.GitRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %w", io.GitRepoURL, err)
	}

	if len(utility.RemoveEmptyStrings(strings.Split(gr.Path, "/"))) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", strings.Trim(gr.Path, ".git"))
	}

	if io.PrivateRepoURLDriver != "" {
		if !supportedDrivers.supported(io.PrivateRepoURLDriver) {
			return fmt.Errorf("invalid driver type: %q", io.PrivateRepoURLDriver)
		}
	}

	if io.SaveTokenKeyRing && io.Secret == "" {
		return errors.New("--secret is required if --save-token-keyring is enabled")
	}
	return nil
}

// Run runs the project Component command.
func (io *BootstrapNewParameters) Run() error {
	log.Progressf("\nCompleting Bootstrap process\n")
	appFs := ioutils.NewFilesystem()
	err := pipelines.BootstrapNew(io.BootstrapNewOptions, appFs)
	if err != nil {
		return err
	}
	if err == nil && io.PushToGit {
		log.Successf("Created repository")
	}
	nextSteps()
	return nil
}

func NewCmdComponent(name, fullName string) *cobra.Command {
	o := NewBootstrapNewParameters()
	var componentCmd = &cobra.Command{
		Use:     name,
		Short:   bootstrapShortDescC,
		Long:    bootstrapLongDescC,
		Example: fmt.Sprintf(bootstrapExampleC, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	componentCmd.Flags().StringVar(&o.Output, "output", "./gitops", "Path to write GitOps resources")
	componentCmd.Flags().StringVar(&o.ComponentName, "component-name", "", "Provide a Component Name within the Application")
	componentCmd.Flags().StringVar(&o.ApplicationName, "application-name", "", "Provide a name for your Application")
	componentCmd.Flags().StringVar(&o.Secret, "secret", "", "Used to authenticate repository clones. Access token is encrypted and stored on local file system by keyring, will be updated/reused.")
	componentCmd.Flags().StringVar(&o.GitRepoURL, "git-repo-url", "", "Provide the URL for your GitOps repository e.g. https://github.com/organisation/repository.git")
	componentCmd.Flags().StringVar(&o.NameSpace, "namespace", "openshift-gitops", "this is a name-space options")
	componentCmd.Flags().IntVar(&o.TargetPort, "target-port", 8080, "Provide the Target Port for your Application")
	componentCmd.Flags().BoolVar(&o.PushToGit, "push-to-git", false, "Overwrites previously existing GitOps configuration (if any) on the local filesystem")
	componentCmd.Flags().StringVar(&o.Route, "route", "", "If you specify the route flag and pass the string, that string will be in the route.yaml that is generated")
	componentCmd.Flags().BoolVar(&o.Interactive, "interactive", false, "If true, enable prompting for most options if not already specified on the command line")
	componentCmd.Flags().BoolVar(&o.Overwrite, "overwrite", false, "If true, it will overwrite the files")
	componentCmd.Flags().BoolVar(&o.SaveTokenKeyRing, "save-token-keyring", false, "Explicitly pass this flag to update the git-host-access-token in the keyring on your local machine")
	componentCmd.Flags().StringVar(&o.PrivateRepoURLDriver, "private-repo-driver", "", "If your Git repositories are on a custom domain, please indicate which driver to use github or gitlab")
	return componentCmd
}

func isKnownDriverURL(repoURL string) bool {
	host, err := accesstoken.HostFromURL(repoURL)
	if err != nil {
		return false
	}
	_, err = factory.DefaultIdentifier.Identify(host)
	return err == nil
}

func nextSteps() {
	log.Success("New Bootstrapped OpenShift resources successfully\n\n",
		"\n",
	)
}
