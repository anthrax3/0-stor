/*
 * Copyright (C) 2017-2018 GIG Technology NV and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pipeline

import (
	"errors"
	"fmt"
	"net"

	clientGRPC "github.com/zero-os/0-stor/client/datastor/grpc"
	serverGRPC "github.com/zero-os/0-stor/server/api/grpc"
	"github.com/zero-os/0-stor/server/db/memory"
)

func newGRPCServerCluster(count int) (*clientGRPC.Cluster, func(), error) {
	if count < 1 {
		return nil, nil, errors.New("invalid GRPC server-client count")
	}
	var (
		cleanupSlice []func()
		addressSlice []string
	)
	for i := 0; i < count; i++ {
		_, addr, cleanup, err := newGRPCServerClient()
		if err != nil {
			for _, cleanup := range cleanupSlice {
				cleanup()
			}
			return nil, nil, err
		}
		cleanupSlice = append(cleanupSlice, cleanup)
		addressSlice = append(addressSlice, addr)
	}

	cluster, err := clientGRPC.NewInsecureCluster(addressSlice, "myLabel", nil)
	if err != nil {
		for _, cleanup := range cleanupSlice {
			cleanup()
		}
		return nil, nil, err
	}

	cleanup := func() {
		cluster.Close()
		for _, cleanup := range cleanupSlice {
			cleanup()
		}
	}
	return cluster, cleanup, nil
}

func newGRPCServerClient() (*clientGRPC.Client, string, func(), error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, "", nil, err
	}

	server, err := serverGRPC.New(memory.New(), serverGRPC.ServerConfig{})
	if err != nil {
		return nil, "", nil, err
	}
	go func() {
		err := server.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	client, err := clientGRPC.NewInsecureClient(listener.Addr().String(), "myLabel", nil)
	if err != nil {
		server.Close()
		return nil, "", nil, err
	}

	clean := func() {
		fmt.Sprintln("clean called")
		err := client.Close()
		if err != nil {
			panic(err)
		}
		err = server.Close()
		if err != nil {
			panic(err)
		}
	}

	return client, listener.Addr().String(), clean, nil
}
