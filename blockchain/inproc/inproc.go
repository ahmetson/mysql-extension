package inproc

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

// Returns the endpoint to the blockchain clients manager.
// Use this endpoint in the req socket to interact with the blockchain nodes.
// The interaction goes through blockchain/<network id>/client.
func ClientEndpoint(network_id string) string {
	return "inproc://spaghetti_" + network_id
}

// Returns the current smartcontract categorizer
// manager url
func CurrentCategorizerEndpoint(network_id string) string {
	return "inproc://cat_current_" + network_id
}

// Returns the old smartcontract categorizer
// manager url
func OldCategorizerEndpoint(network_id string) string {
	return "inproc://cat_old_" + network_id
}

// Returns the categorizer manager url
func CategorizerEndpoint(network_id string) string {
	return "inproc://cat_" + network_id
}

func CurrentCategorizerManagerSocket(network_id string) (*zmq.Socket, error) {
	sock, err := zmq.NewSocket(zmq.PUSH)
	if err != nil {
		return nil, fmt.Errorf("zmq error for new push socket: %w", err)
	}

	url := CurrentCategorizerEndpoint(network_id)
	if err := sock.Bind(url); err != nil {
		return nil, fmt.Errorf("trying to create categorizer for network id %s: %v", network_id, err)
	}

	return sock, nil
}

func OldCategorizerManagerSocket(network_id string) (*zmq.Socket, error) {
	sock, err := zmq.NewSocket(zmq.PUSH)
	if err != nil {
		return nil, fmt.Errorf("zmq error for new push socket: %w", err)
	}

	url := OldCategorizerEndpoint(network_id)
	if err := sock.Bind(url); err != nil {
		return nil, fmt.Errorf("trying to create categorizer for network id %s: %v", network_id, err)
	}

	return sock, nil
}

func CategorizerManagerSocket(network_id string) (*zmq.Socket, error) {
	sock, err := zmq.NewSocket(zmq.PUSH)
	if err != nil {
		return nil, fmt.Errorf("zmq error for new push socket: %w", err)
	}

	url := CategorizerEndpoint(network_id)
	if err := sock.Bind(url); err != nil {
		return nil, fmt.Errorf("trying to create categorizer for network id %s: %v", network_id, err)
	}

	return sock, nil
}
