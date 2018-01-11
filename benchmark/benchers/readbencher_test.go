package benchers

import (
	"math"
	"testing"

	"github.com/zero-os/0-stor/benchmark/config"

	"github.com/stretchr/testify/require"
)

func TestReadBencherRuns(t *testing.T) {
	require := require.New(t)

	// setup test servers
	servers, cleanupZstor := newTestZstorServers(t, 4)
	defer cleanupZstor()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Address()
	}

	clientConfig := newDefaultZstorConfig(shards, nil, 64)

	const runs = 5
	sc := config.Scenario{
		ZstorConf: clientConfig,
		BenchConf: config.BenchmarkConfig{
			Method:     "read",
			Operations: runs,
			KeySize:    5,
			ValueSize:  25,
		},
	}

	// run limited benchmark
	rb, err := NewReadBencher(testID, &sc)
	require.NoError(err)

	res, err := rb.RunBenchmark()
	require.NoError(err)
	require.Equal(int64(runs), res.Count)
}

func TestReadBencherDuration(t *testing.T) {
	require := require.New(t)

	// setup test servers
	servers, cleanupZstor := newTestZstorServers(t, 4)
	defer cleanupZstor()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Address()
	}

	clientConfig := newDefaultZstorConfig(shards, nil, 64)

	sc := config.Scenario{
		ZstorConf: clientConfig,
		BenchConf: config.BenchmarkConfig{
			Method:    "read",
			Duration:  duration,
			KeySize:   5,
			ValueSize: 25,
			Output:    "per_second",
		},
	}

	// run limited benchmark
	rb, err := NewReadBencher(testID, &sc)
	require.NoError(err)

	r, err := rb.RunBenchmark()
	require.NoError(err)

	// check if it ran for about requested duration
	runDur := r.Duration.Seconds()
	require.Equal(float64(duration), math.Floor(runDur),
		"rounded run duration should be equal to the requested duration")
}
