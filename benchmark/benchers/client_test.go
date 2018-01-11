package benchers

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testKey   = []byte("testKey")
	testValue = []byte("testValue")
)

func TestInMemoryMetaClient(t *testing.T) {
	require := require.New(t)

	servers, cleanupZstor := newTestZstorServers(t, 4)
	defer cleanupZstor()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Address()
	}

	clientConfig := newDefaultZstorConfig(shards, nil, 64)

	client, err := newClientFromConfig(&clientConfig, 1, false)
	require.NoError(err, "Failed to create client")

	_, err = client.Write(testKey, bytes.NewReader(testValue))
	require.NoError(err, "Failed to write to client")

	buf := bytes.NewBuffer(nil)
	err = client.Read(testKey, buf)
	require.NoError(err, "Failed to read from client")
	require.Equal(testValue, buf.Bytes(), "Read value should be equal to value originally set in the zstor")
}
