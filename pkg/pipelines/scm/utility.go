package scm

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/jenkins-x/go-scm/scm/factory"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var (
	branchRefOverlay = []triggersv1.CELOverlay{
		{Key: "ref", Expression: "body.ref.split('/')[2]"},
	}
)

func invalidRepoPathError(gitType, path string) error {
	return fmt.Errorf("invalid repository path for %s: %s", gitType, path)
}

func unsupportedGitTypeError(gitType string) error {
	return fmt.Errorf("unsupported Git repository type: %s", gitType)
}

func invalidRepoURLError(repoURL, reason string) error {
	return fmt.Errorf("invalid repository URL %s: %s", repoURL, reason)
}

func createEventInterceptor(filter, repoName string) (*triggersv1.EventInterceptor, error) {
	rawFilter, rawOverlays, err := celParams(filter, repoName)
	if err != nil {
		return nil, err
	}
	return &triggersv1.EventInterceptor{
		Ref: triggersv1.InterceptorRef{
			Name: "cel",
		},
		Params: []triggersv1.InterceptorParams{
			{
				Name: "filter",
				Value: v1.JSON{
					Raw: rawFilter,
				},
			},
			{
				Name: "overlays",
				Value: v1.JSON{
					Raw: rawOverlays,
				},
			},
		},
	}, nil
}

func createListenerTemplate(name *string) *triggersv1.EventListenerTemplate {
	return &triggersv1.EventListenerTemplate{
		Ref: name,
	}
}

func createListenerBinding(name string) *triggersv1.EventListenerBinding {
	return &triggersv1.EventListenerBinding{
		Ref: name,
	}
}

func createBindings(names []string) []*triggersv1.EventListenerBinding {
	bindings := make([]*triggersv1.EventListenerBinding, len(names))
	for i, name := range names {
		bindings[i] = createListenerBinding(name)
	}
	return bindings
}

func createBindingParam(name, value string) triggersv1.Param {
	return triggersv1.Param{
		Name:  name,
		Value: value,
	}
}

func processRawURL(rawURL string, processPath func(*url.URL) (string, error)) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	path, err := processPath(parsedURL)
	if err != nil {
		return "", err
	}
	return path, nil
}

func splitRepositoryPath(parsedURL *url.URL) ([]string, error) {
	var components []string
	for _, s := range strings.Split(parsedURL.Path, "/") {
		if s != "" {
			components = append(components, s)
		}
	}
	if len(components) < 1 {
		return nil, invalidRepoURLError(parsedURL.String(), "path is empty")
	}
	components[len(components)-1] = strings.TrimSuffix(components[len(components)-1], ".git")
	return components, nil
}

// GetDriverName gets the driver to be used for this repo url, using the go-scm
// default identifier.
func GetDriverName(rawURL string) (string, error) {
	host, err := HostnameFromURL(rawURL)
	if err != nil {
		return "", err
	}
	return factory.DefaultIdentifier.Identify(host)
}

// HostnameFromURL returns the host from a URL.
func HostnameFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return strings.ToLower(u.Host), nil
}

func secretParam(name, key string) ([]byte, error) {
	return json.Marshal(map[string]string{
		"secretName": name,
		"secretKey":  key,
	})
}

func celParams(filter, repoName string) ([]byte, []byte, error) {
	rawFilter, err := json.Marshal(fmt.Sprintf(filter, repoName))
	if err != nil {
		return nil, nil, err
	}
	rawOverlays, err := json.Marshal(branchRefOverlay)
	if err != nil {
		return nil, nil, err
	}
	return rawFilter, rawOverlays, nil
}

func eventInterceptorWithSecret(name string, secretInfo []byte) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		Ref: triggersv1.InterceptorRef{
			Name: name,
		},
		Params: []triggersv1.InterceptorParams{
			{
				Name: "secretRef",
				Value: v1.JSON{
					Raw: secretInfo,
				},
			},
		},
	}
}
