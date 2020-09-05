package utility

import (
	"strings"

	"github.com/openshift/odo/pkg/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/kubernetes"
)

// AddGitSuffixIfNecessary will append .git to URL if necessary
func AddGitSuffixIfNecessary(url string) string {
	if url == "" || strings.HasSuffix(strings.ToLower(url), ".git") {
		return url
	}
	log.Infof("Adding .git to %s", url)
	return url + ".git"
}

// RemoveEmptyStrings returns a slice with all the empty strings removed from the
// source slice.
func RemoveEmptyStrings(s []string) []string {
	nonempty := []string{}
	for _, v := range s {
		if v != "" {
			nonempty = append(nonempty, v)
		}
	}
	return nonempty
}

// MaybeCompletePrefix adds a hyphen on the end of the prefix if it doesn't have
// one to make prefix-generated names look a bit nicer.
func MaybeCompletePrefix(s string) string {
	if s != "" && !strings.HasSuffix(s, "-") {
		return s + "-"
	}
	return s
}

// Client represents a client for K8s
type Client struct {
	KubeClient kubernetes.Interface
}

// NewClient returns a new K8s client
func NewClient(client kubernetes.Interface) *Client {
	return &Client{KubeClient: client}
}

// CheckIfSealedSecretsExists checks if sealed secrets is installed
func (c *Client) CheckIfSealedSecretsExists(secret types.NamespacedName) error {
	_, err := c.KubeClient.CoreV1().Services(secret.Namespace).Get(secret.Name, v1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

// CheckIfArgoCDExists checks if ArgoCD operator is installed
func (c *Client) CheckIfArgoCDExists(ns string) error {
	_, err := c.KubeClient.AppsV1().Deployments(ns).Get("argocd-operator", v1.GetOptions{})
	if err != nil {
		return err
	}

	// check if ArgoCD instance is created
	_, err = c.KubeClient.AppsV1().Deployments(ns).Get("argocd-server", v1.GetOptions{})
	if err != nil {
		return err
	}

	return err
}

// CheckIfPipelinesExists checks is OpenShift pipelines operator is installed
func (c *Client) CheckIfPipelinesExists(ns string) error {
	_, err := c.KubeClient.AppsV1().Deployments(ns).Get("openshift-pipelines-operator", v1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}
