package config

import (
	"io/ioutil"
	"math"
	"testing"

	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/client/itsyouonline"

	"github.com/stretchr/testify/require"
)

const (
	validFile              = "../../fixtures/benchmark/testconfigs/valid_conf.yaml"
	emptyFile              = "../../fixtures/benchmark/testconfigs/empty_conf.yaml"
	invalidDurOpsConfFile  = "../../fixtures/benchmark/testconfigs/invalid_dur_ops_conf.yaml"
	invalidKeySizeConfFile = "../../fixtures/benchmark/testconfigs/invalid_keysize_conf.yaml"
)

func TestClientConfig(t *testing.T) {
	require := require.New(t)

	yamlFile, err := ioutil.ReadFile(validFile)
	require.NoError(err)

	_, err = UnmarshalYAML(yamlFile)
	require.NoError(err)
}

func TestInvalidClientConfig(t *testing.T) {
	require := require.New(t)

	// empty config
	yamlFile, err := ioutil.ReadFile(emptyFile)
	require.NoError(err)

	_, err = UnmarshalYAML(yamlFile)
	require.Error(err)

	// invalid ops/duration
	yamlFile, err = ioutil.ReadFile(invalidDurOpsConfFile)
	require.NoError(err)

	_, err = UnmarshalYAML(yamlFile)
	require.Error(err)

	// invalid keysize
	yamlFile, err = ioutil.ReadFile(invalidKeySizeConfFile)
	require.NoError(err)

	_, err = UnmarshalYAML(yamlFile)
	require.Error(err)
}

func TestSetupClientConfig(t *testing.T) {
	require := require.New(t)
	c := client.Config{
		IYO: itsyouonline.Config{
			Organization:      "org",
			ApplicationID:     "some ID",
			ApplicationSecret: "some secret",
		},
	}

	SetupClientConfig(&c)
	require.NotEmpty(c.Namespace, "Namespace should be set")

	const testNamespace = "test_namespace"
	c = client.Config{
		IYO: itsyouonline.Config{
			Organization:      "org",
			ApplicationID:     "some ID",
			ApplicationSecret: "some secret",
		},
		Namespace: testNamespace,
	}

	SetupClientConfig(&c)
	require.Equal(testNamespace, c.Namespace)
}

func TestMarshalBencherMethod(t *testing.T) {
	require := require.New(t)

	testCases := []struct {
		Type     BencherMethod
		Expected string
	}{
		{BencherRead, "read"},
		{BencherWrite, "write"},
		{math.MaxUint8, ""},
	}
	for _, tc := range testCases {
		b, err := tc.Type.MarshalText()
		if tc.Expected == "" {
			require.Error(err)
			require.Nil(b)
		} else {
			require.NoError(err)
			require.Equal(tc.Expected, string(b))
		}
	}
}

func TestUnmarshalBencherMethod(t *testing.T) {
	require := require.New(t)

	testCases := []struct {
		String   string
		Expected BencherMethod
		Err      bool
	}{
		{"read", BencherRead, false},
		{"Read", BencherRead, false},
		{"READ", BencherRead, false},
		{"write", BencherWrite, false},
		{"Write", BencherWrite, false},
		{"WRITE", BencherWrite, false},
		{"foo", math.MaxUint8, true},
	}
	for _, tc := range testCases {
		var bm BencherMethod
		err := bm.UnmarshalText([]byte(tc.String))
		if tc.Err {
			require.Error(err)
		} else {
			require.NoError(err)
			require.Equal(tc.Expected, bm)
		}
	}
}
