package inproc

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

func NewCategorizerPusher(network_id string) (*zmq.Socket, error) {
	sock, err := zmq.NewSocket(zmq.PUSH)
	if err != nil {
		return nil, err
	}

	url := "cat_" + network_id
	if err := sock.Connect("inproc://" + url); err != nil {
		return nil, fmt.Errorf("trying to create categorizer for network id %s: %v", network_id, err)
	}

	return sock, nil
}
