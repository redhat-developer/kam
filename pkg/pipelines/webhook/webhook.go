package webhook

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/redhat-developer/kam/pkg/pipelines/config"
	"github.com/redhat-developer/kam/pkg/pipelines/eventlisteners"
	"github.com/redhat-developer/kam/pkg/pipelines/git"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/redhat-developer/kam/pkg/pipelines/routes"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"
	"github.com/zalando/go-keyring"
)

type webhookInfo struct {
	clusterResource *resources
	repository      *git.Repository
	gitRepoURL      string
	cicdNamepace    string
	listenerURL     string
	accessToken     string
	serviceName     *QualifiedServiceName
	isCICD          bool
}

// KeyringServiceName refers to service name used to set the accesstoken in the keyring
const KeyringServiceName = "kam"

// QualifiedServiceName represents three part name of a service (Environment, Application, and Service)
type QualifiedServiceName struct {
	EnvironmentName string
	ServiceName     string
}

// Create creates a new webhook on the target Git Repository
// It returns the ID of created webhook.
func Create(accessToken, pipelinesFile string, serviceName *QualifiedServiceName, isCICD bool) (string, error) {
	webhook, err := newWebhookInfo(accessToken, pipelinesFile, serviceName, isCICD)
	if err != nil {
		return "", err
	}

	exists, err := webhook.exists()
	if err != nil {
		return "", err
	}

	if exists {
		return "", errors.New("webhook already exists")
	}

	return webhook.create()
}

// Delete deletes webhooks on the target Git Repository that match the listener address
// It returns the IDs of deleted webhooks.
func Delete(accessToken, pipelinesFile string, serviceName *QualifiedServiceName, isCICD bool) ([]string, error) {
	webhook, err := newWebhookInfo(accessToken, pipelinesFile, serviceName, isCICD)
	if err != nil {
		return nil, err
	}

	ids, err := webhook.list()
	if err != nil {
		return nil, err
	}

	return webhook.delete(ids)
}

// List returns an array of webhook IDs for the target Git repository/listeners
func List(accessToken, pipelinesFile string, serviceName *QualifiedServiceName, isCICD bool) ([]string, error) {
	webhook, err := newWebhookInfo(accessToken, pipelinesFile, serviceName, isCICD)
	if err != nil {
		return nil, err
	}

	return webhook.list()
}

func newWebhookInfo(accessToken, pipelinesFile string, serviceName *QualifiedServiceName, isCICD bool) (*webhookInfo, error) {
	manifest, err := config.LoadManifest(ioutils.NewFilesystem(), pipelinesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pipelines: %v", err)
	}

	gitRepoURL := getRepoURL(manifest, isCICD, serviceName)
	if gitRepoURL == "" {
		return nil, errors.New("failed to find Git repository URL in manifest")
	}

	cfg := manifest.GetPipelinesConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get CICD environment: %v", err)
	}
	cicdNamepace := cfg.Name

	clusterResources, err := newResources()
	if err != nil {
		return nil, err
	}

	listenerURL, err := getListenerURL(clusterResources, cicdNamepace)
	if err != nil {
		return nil, fmt.Errorf("failed to get event listener URL: %v", err)
	}
	if accessToken == "" {
		accessToken, err = GetAccessToken(gitRepoURL)
		if err != nil {
			return nil, err
		}
		if accessToken == "" && err == nil {
			return nil, errors.New("unable to retrieve access-token from keyring/environment variables")
		}
	} else {
		err := SetSecret(gitRepoURL, accessToken)
		if err != nil {
			return nil, err
		}

	}
	repository, err := git.NewRepository(gitRepoURL, accessToken)
	if err != nil {
		return nil, err
	}
	return &webhookInfo{clusterResources, repository, gitRepoURL, cicdNamepace, listenerURL, accessToken, serviceName, isCICD}, nil
}

// GetAccessToken returns the token from either that is stored in the keyring or the environment variable in this order.
func GetAccessToken(gitRepoURL string) (string, error) {
	hostName, err := HostFromURL(gitRepoURL)
	if err != nil {
		return "", err
	}
	accessToken, err := keyring.Get(KeyringServiceName, hostName)
	if err != nil && err != keyring.ErrNotFound {
		return "", err
	}
	if err != nil && err == keyring.ErrNotFound {
		FmtHostName := strings.ReplaceAll(hostName, ".", "_")
		envVar := strings.ToUpper(FmtHostName) + "_TOKEN"
		accessToken = os.Getenv(envVar)
		if accessToken == "" {
			return "", nil
		}
	}
	return accessToken, nil
}

func (w *webhookInfo) exists() (bool, error) {
	ids, err := w.repository.ListWebhooks(w.listenerURL)
	if err != nil {
		return false, err
	}

	return len(ids) > 0, nil
}

func (w *webhookInfo) list() ([]string, error) {
	return w.repository.ListWebhooks(w.listenerURL)
}

func (w *webhookInfo) delete(ids []string) ([]string, error) {
	return w.repository.DeleteWebhooks(ids)
}

func (w *webhookInfo) create() (string, error) {
	secret, err := getWebhookSecret(w.clusterResource, w.cicdNamepace, w.isCICD, w.serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to get webhook secret: %v", err)
	}

	return w.repository.CreateWebhook(w.listenerURL, secret)
}

// Get Git repository URL whether it is CICD configuration or service source repository
// Return "" if not found
func getRepoURL(manifest *config.Manifest, isCICD bool, serviceName *QualifiedServiceName) string {
	if isCICD {
		return manifest.GitOpsURL
	}

	return getSourceRepoURL(manifest, serviceName)
}

// Get service source repository URL.  Return "" if not found
func getSourceRepoURL(manifest *config.Manifest, service *QualifiedServiceName) string {
	for _, env := range manifest.Environments {
		if env.Name == service.EnvironmentName {
			for _, app := range env.Apps {
				for _, svc := range app.Services {
					if svc.Name == service.ServiceName {
						return svc.SourceURL
					}
				}
			}
		}
	}
	return ""
}

func getListenerURL(r *resources, cicdNamespace string) (string, error) {
	hasTLS, host, err := r.getListenerAddress(cicdNamespace, routes.GitOpsWebhookEventListenerRouteName)
	if err != nil {
		return "", err
	}

	return buildURL(host, hasTLS), nil
}

func buildURL(host string, hasTLS bool) string {
	scheme := "http"
	if hasTLS {
		scheme += "s"
	}

	return scheme + "://" + host
}

func getWebhookSecret(r *resources, namespace string, isCICD bool, service *QualifiedServiceName) (string, error) {
	var secretName string
	if isCICD {
		secretName = eventlisteners.GitOpsWebhookSecret
	} else {
		// currently, use the app name to create webhook secret name.
		// also currently, service webhook secret are in CICI namespace
		secretName = secrets.MakeServiceWebhookSecretName(service.EnvironmentName, service.ServiceName)
	}
	return r.getWebhookSecret(namespace, secretName, eventlisteners.WebhookSecretKey)
}

// HostFromURL extracts the hostname from the url passed
func HostFromURL(s string) (string, error) {
	p, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return strings.ToLower(p.Host), nil
}

//SetSecret sets the secret in the keyring
func SetSecret(repoURL, accessToken string) error {
	hostName, err := HostFromURL(repoURL)
	if err != nil {
		return err
	}
	secret, err := getSecret(hostName)
	if err != nil {
		return err
	}
	if accessToken != secret {
		err := keyring.Set(KeyringServiceName, hostName, accessToken)
		if err != nil {
			return fmt.Errorf("unable to set access token for repo %q using keyring: %w", repoURL, err)
		}
	}
	return nil
}

func getSecret(hostName string) (string, error) {
	secret, err := keyring.Get(KeyringServiceName, hostName)
	if err != nil && err != keyring.ErrNotFound {
		return "", err
	}
	return secret, nil
}
