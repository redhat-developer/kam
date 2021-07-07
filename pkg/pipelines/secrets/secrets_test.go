package secrets

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/kam/test"
)

func init() {
	testModulus = new(big.Int)
	_, err := fmt.Sscan("777304254876434297689544225447769213262492599515515837291621795936355252933930193245809942636192119684040605554803489669141565417296821660595336672178414512660751886699171738066307588619202437848899334837760648051656982184646490661921128886671800776058692981991859399404705935722225294811424879738586269551402668122524371718537515440568440102201259925611463161144897905846190044735554045001999198442528435295995584980713050916813579912296878368079243909549993116827192901474611239264189340401059113919551426849847211275352102674049634252149163111599977742365280992561904350781270344655927564475032580504276518647106167707150111291732645399166011800154961975117045723373023335778593638216165426988399138193230056486079421256484837299169853958601000282124667227789126483641999102102039577368681983584245367307077546423870452524154641890843463963116237003367269116435430641427113406369059991147359641266708862913786891945896441771663010146473536372286482453315017377528517965715554550898957321536181165129538808789201530141159181590893764287807749414277289452691723903046140558704697831351834538780165261072894792900501671534138992265545905216973214953125367388406669893889742303072755608685449114438926280862339744991872488262084141163", testModulus)
	if err != nil {
		panic(err)
	}
}

const (
	testToken = "abcdefghijklmnop"
)

var (
	testModulus *big.Int
)

func TestCreateOpaqueSecret(t *testing.T) {
	data := "abcdefghijklmnop"
	secret, err := createOpaqueSecret(meta.NamespacedName("cicd", "github-auth"), data, "token")
	if err != nil {
		t.Fatal(err)
	}

	want := &corev1.Secret{
		TypeMeta: secretTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "github-auth",
			Namespace: "cicd",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"token": []byte(data),
		},
	}

	if diff := cmp.Diff(want, secret); diff != "" {
		t.Fatalf("createOpaqueSecret() failed got\n%s", diff)
	}
}

func TestCreateDockerConfigSecretWithErrorReading(t *testing.T) {
	testErr := errors.New("test failure")
	_, err := createDockerConfigSecret(meta.NamespacedName("cici", "github-auth"), errorReader{testErr})
	test.AssertErrorMatch(t, "failed to read .* test failure", err)
}

func TestCreateDockerConfigSecret(t *testing.T) {
	data := []byte(testToken)
	secret, err := createDockerConfigSecret(meta.NamespacedName("cicd", "regcred"), bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	want := &corev1.Secret{
		TypeMeta: secretTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "regcred",
			Namespace: "cicd",
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": data,
		},
	}

	if diff := cmp.Diff(want, secret); diff != "" {
		t.Fatalf("createDockerConfigSecret() failed got\n%s", diff)
	}
}

func TestBasicAuthSecret(t *testing.T) {
	host := "https://github.com"
	secret := createBasicAuthSecret(meta.NamespacedName("cicd", "github-auth"), testToken, meta.AddAnnotations(
		map[string]string{
			"tekton.dev/git-0": host,
		}),
	)

	want := &corev1.Secret{
		TypeMeta: secretTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "github-auth",
			Namespace: "cicd",
			Annotations: map[string]string{
				"tekton.dev/git-0": host,
			},
		},
		Type: corev1.SecretTypeBasicAuth,
		StringData: map[string]string{
			"username": "tekton",
			"password": testToken,
		},
	}

	if diff := cmp.Diff(want, secret); diff != "" {
		t.Fatalf("createBasicAuthSecret() failed got\n%s", diff)
	}
}

type errorReader struct {
	err error
}

func (e errorReader) Read(p []byte) (int, error) {
	return 0, e.err
}
