// the categorizer package keeps data types used by SDS Categorizer.
// the data type functions as well as method to obtain data from SDS Categorizer.
package abi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	static_abi "github.com/blocklords/sds/static/abi"
	eth_common "github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// //////////////////////////////////////////////////////////////////////////
//
// Abi struct is used for EVM based categorizer.
// it has the smartcontract interface to parse the raw spaghetti data into categorized data.
// its the wrapper over the SDS Static abi.
//
// //////////////////////////////////////////////////////////////////////////
type Abi struct {
	static_abi *static_abi.Abi
	geth_abi   abi.ABI // interface
}

// Returns an abi.Method from geth
func (a *Abi) GetMethod(method string) (*abi.Method, error) {
	for _, m := range a.geth_abi.Methods {
		if m.Name == method {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("method %s not found in abi", method)
}

// Given the transaction data, returns a categorized variant.
//
// The first returning parameter is the method name, second parameter are the method arguments as map of
// argument name => value
func (a *Abi) Categorize(data string) (string, map[string]interface{}, error) {
	inputs := map[string]interface{}{}

	offset := 0
	if len(data) > 2 && data[:2] == "0x" || data[:2] == "0X" {
		offset += 2
	}

	// decode method signature
	sig, err := hex.DecodeString(data[offset : 8+offset])
	if err != nil {
		return "", inputs, fmt.Errorf("failed to extract method signature from transaction data. the hex package error: %w", err)
	}

	// recover Method from signature and ABI
	method, err := a.geth_abi.MethodById(sig)
	if err != nil {
		return "", inputs, fmt.Errorf("failed to find a method by its signature. geth package error: %w", err)
	}

	// decode txInput Payload
	decoded_data, err := hex.DecodeString(data[8+offset:])
	if err != nil {
		return method.Name, inputs, fmt.Errorf("failed to extract method input arguments from transaction data. the hex package error: %w", err)
	}

	err = method.Inputs.UnpackIntoMap(inputs, decoded_data)
	if err != nil {
		return method.Name, inputs, fmt.Errorf("failed to parse method input parameters into map. the geth package error: %w", err)
	}

	return method.Name, inputs, nil
}

// it adds an ethereum abi layer on top of the static abi
func NewAbi(static_abi *static_abi.Abi) (*Abi, error) {
	abi_obj := Abi{static_abi: static_abi}

	if err := json.Unmarshal(static_abi.Bytes, &abi_obj.geth_abi); err != nil {
		return nil, fmt.Errorf("failed to decompose abi to geth abi: %w", err)
	}

	return &abi_obj, nil
}

func get_indexed(inputs abi.Arguments) abi.Arguments {
	ret := make(abi.Arguments, 0)
	for _, arg := range inputs {
		if arg.Indexed {
			ret = append(ret, arg)
		}
	}
	return ret
}

func (a *Abi) DecodeLog(topics []string, data string) (string, map[string]interface{}, error) {
	if len(topics) == 0 {
		return "", nil, fmt.Errorf("anonymous events are not supported")
	}

	topic_hashes := make([]eth_common.Hash, len(topics)-1)
	var event_id eth_common.Hash
	for i, topic := range topics {
		if i == 0 {
			event_id = eth_common.HexToHash(topic)
		} else {
			topic_hashes[i-1] = eth_common.HexToHash(topic)
		}
	}

	topic_outputs := make(map[string]interface{}, 0)

	data_outputs := make(map[string]interface{}, 0)
	for _, event := range a.geth_abi.Events {
		if strings.EqualFold(event_id.String(), event.ID.String()) {
			if len(data) > 0 {
				bytes, err := hex.DecodeString(data)
				if err != nil {
					return "", nil, fmt.Errorf("error decoding data strin to bytes: %w", err)
				}
				err = event.Inputs.NonIndexed().UnpackIntoMap(data_outputs, bytes)
				if err != nil {
					return "", nil, fmt.Errorf("parsing event %s for data %s error: %w", event.RawName, bytes, err)
				}
			}

			indexed := get_indexed(event.Inputs)
			err := abi.ParseTopicsIntoMap(topic_outputs, indexed, topic_hashes)
			if err != nil {
				return "", nil, fmt.Errorf("event %s for %v topics parsing error: %w", event.RawName, topics, err)
			}

			// merge topics and data
			for key, value := range topic_outputs {
				data_outputs[key] = value
			}

			return event.RawName, data_outputs, nil
		}
	}

	return "", nil, fmt.Errorf("failed to decode the event. No matching signature")
}
