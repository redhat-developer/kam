package utility

import (
	"strings"

	"github.com/openshift/odo/pkg/log"
	operatorsclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/clientconfig"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	argocdCRD = "argocds.argoproj.io"
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
	KubeClient     kubernetes.Interface
	OperatorClient operatorsclientset.OperatorsV1alpha1Interface
}

// NewClient returns a new client to check dependencies
func NewClient() (*Client, error) {
	clientConfig, err := clientconfig.GetRESTConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	operatorClientSet, err := operatorsclientset.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	return &Client{KubeClient: clientSet, OperatorClient: operatorClientSet}, nil
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
	csvList, err := c.OperatorClient.ClusterServiceVersions(ns).List(v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, csv := range csvList.Items {
		if csv.OwnsCRD(argocdCRD) {
			return nil
		}
	}
	return errors.NewNotFound(schema.GroupResource{Group: "argoproj.io", Resource: "argocd"}, argocdCRD)
}

// CheckIfPipelinesExists checks is OpenShift pipelines operator is installed
func (c *Client) CheckIfPipelinesExists(ns string) error {
	_, err := c.KubeClient.AppsV1().Deployments(ns).Get("openshift-pipelines-operator", v1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

// GetFullName generates a command's full name based on its parent's full name and its own name
func GetFullName(parentName, name string) string {
	return parentName + " " + name
}
