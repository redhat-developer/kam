package secrets

import (
	"crypto/rsa"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/redhat-developer/kam/pkg/pipelines/meta"
)

var (
	secretTypeMeta = meta.TypeMeta("Secret", "v1")
)

// PublicKeyFunc retruns a public key  give a service namedspaced name
type PublicKeyFunc func(service types.NamespacedName) (*rsa.PublicKey, error)

// MakeServiceWebhookSecretName common method to create service webhook secret name
func MakeServiceWebhookSecretName(envName, serviceName string) string {
	return fmt.Sprintf("webhook-secret-%s-%s", envName, serviceName)
}

// CreateUnsealedDockerConfigSecret creates an Unsealed Secret with the given name and reader
func CreateUnsealedDockerConfigSecret(name types.NamespacedName, in io.Reader) (*corev1.Secret, error) {
	secret, err := createDockerConfigSecret(name, in)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func CreateUnsealedSecret(name types.NamespacedName, data, secretKey string) (*corev1.Secret, error) {
	secret, err := createOpaqueSecret(name, data, secretKey)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// CreateUnsealedBasicAuthSecret creates a SealedSecret with a BasicAuth type
// secret.
func CreateUnsealedBasicAuthSecret(name types.NamespacedName, token string,
	opts ...meta.ObjectMetaOpt) *corev1.Secret {
	return createBasicAuthSecret(name, token, opts...)
}

// createOpaqueSecret creates a Kubernetes v1/Secret with the provided name and
// body, and type Opaque.
func createOpaqueSecret(name types.NamespacedName, data, secretKey string) (*corev1.Secret, error) {
	r := strings.NewReader(data)
	return createSecret(name, secretKey, corev1.SecretTypeOpaque, r)
}

// createDockerConfigSecret creates a Kubernetes v1/Secret with the provided name and
// body, and type DockerConfigJson.
func createDockerConfigSecret(name types.NamespacedName, in io.Reader) (*corev1.Secret, error) {
	return createSecret(name, ".dockerconfigjson", corev1.SecretTypeDockerConfigJson, in)
}

func createBasicAuthSecret(name types.NamespacedName, token string, opts ...meta.ObjectMetaOpt) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta:   secretTypeMeta,
		ObjectMeta: meta.ObjectMeta(name, opts...),
		Type:       corev1.SecretTypeBasicAuth,
		StringData: map[string]string{
			"username": "tekton",
			"password": token,
		},
	}
}

func createSecret(name types.NamespacedName, key string, st corev1.SecretType, in io.Reader) (*corev1.Secret, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret data: %v", err)
	}
	secret := &corev1.Secret{
		TypeMeta:   secretTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Type:       st,
		Data: map[string][]byte{
			key: data,
		},
	}
	return secret, nil
}
