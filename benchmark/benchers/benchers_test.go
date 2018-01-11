package benchers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBencherCtor(t *testing.T) {
	require := require.New(t)
	validBenchers := []string{
		"read",
		"write",
		"Read",
		"Write",
		"READ",
		"WRITE",
	}

	for _, benchMethod := range validBenchers {
		bencher := GetBencherCtor(benchMethod)
		require.NotNil(bencher, "Valid bencher should not be nil")
	}

	invalidBenchers := []string{
		"reading",
		"writed",
		"Lorem Ipsum",
		"",
	}

	for _, benchMethod := range invalidBenchers {
		bencher := GetBencherCtor(benchMethod)
		require.Nil(bencher, "Invalid bencher should be nil")
	}
}
