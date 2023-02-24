package network_client

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/blocklords/gosds/spaghetti/block"
	"github.com/blocklords/gosds/spaghetti/transaction"

	"github.com/ethereum/go-ethereum"

	"github.com/blocklords/gosds/static/network"

	eth_common "github.com/ethereum/go-ethereum/common"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NetworkClient struct {
	network_id string
	client     *ethclient.Client
	ctx        context.Context
	Network    *network.Network
}

// Create a network client connected to the blockchain based on a Static parameters
// Static parameters include the node url
func New(network *network.Network) (*NetworkClient, error) {
	ctx := context.TODO()
	client, err := ethclient.DialContext(ctx, network.Provider)
	if err != nil {
		return nil, errors.New(`failed address connect address the provider. please try again later. error from provider package: ` + err.Error())
	}

	return &NetworkClient{
		client:  client,
		ctx:     ctx,
		Network: network,
	}, nil
}

// Creates a network clients connected to the blockchain network for each static parameter
func SetupClients(networks []*network.Network) (map[string]*NetworkClient, error) {
	network_clients := make(map[string]*NetworkClient, len(networks))

	for _, network := range networks {
		new_client, err := New(network)
		if err != nil {
			return nil, errors.New(err.Error())
		}

		network_clients[network.Id] = new_client
	}

	return network_clients, nil
}

//////////////////////////////////////////////////////////
//
// Blockchain related functions
//
/////////////////////////////////////////////////////////

// Returns the block timestamp from the blockchain
func (c *NetworkClient) GetBlockTimestamp(block_number uint64) (uint64, error) {
	header, err := c.client.HeaderByNumber(c.ctx, big.NewInt(int64(block_number)))
	if err != nil {
		return 0, errors.New("failed to fetch block information from blockchain: " + err.Error())
	}

	return header.Time, nil
}

// Returns the most recent block number from blockchain
func (c *NetworkClient) GetRecentBlockNumber() (uint64, error) {
	return c.client.BlockNumber(c.ctx)
}

// Returns the information about the specific transaction from the blockchain
// The transaction is converted into the gosds/spaghetti/transaction.Transaction data type
func (c *NetworkClient) GetTransaction(transaction_id string) (*transaction.Transaction, error) {
	transaction_hash := eth_common.HexToHash(transaction_id)
	var transaction_raw *eth_types.Transaction
	var pending bool
	var err error

	attempt := 10
	for {
		transaction_raw, pending, err = c.client.TransactionByHash(c.ctx, transaction_hash)
		if pending {
			return nil, fmt.Errorf("the transaction is in the pending mode. please try again later fetching %s", transaction_hash)
		}
		if err == nil {
			break
		}
		if attempt == 0 {
			return nil, fmt.Errorf("transaction by hash error after 10 attempts: " + err.Error())
		}
		fmt.Println("transaction by hash wasn't found for txid ", transaction_id, "at network", c.network_id, " retrying again")
		time.Sleep(time.Second * 1)
		attempt--
	}

	var receipt *eth_types.Receipt
	attempt = 10
	for {
		receipt, err = c.client.TransactionReceipt(c.ctx, transaction_hash)
		if err == nil {
			break
		}
		if attempt == 0 {
			return nil, fmt.Errorf("transaction receipt error after 10 attempts: " + err.Error())
		}
		fmt.Println("transaction by receipt wasn't found for txid ", transaction_hash, " at network ", c.network_id, " retrying again")
		time.Sleep(time.Second * 1)
		attempt--
	}

	tx, parse_err := transaction.New(c.network_id, receipt.BlockNumber.Uint64(), receipt.TransactionIndex, transaction_raw)
	if parse_err != nil {
		return nil, parse_err
	}
	if tx.TxTo == "" {
		tx.TxTo = receipt.ContractAddress.Hex()
	}

	return tx, nil
}

// Returns the block with transactions and logs
func (c *NetworkClient) GetBlock(block_number uint64) (*block.Block, error) {
	big_int := big.NewInt(int64(block_number))

	raw_block, err := c.client.BlockByNumber(c.ctx, big_int)
	if err != nil {
		return nil, err
	}
	b := &block.Block{
		NetworkId:      c.Network.Id,
		BlockNumber:    raw_block.NumberU64(),
		BlockTimestamp: raw_block.Time(),
		Logs:           nil,
	}

	var raw_logs []eth_types.Log
	var log_err error
	attempt := 5
	for {
		raw_logs, log_err = c.GetBlockLogs(block_number)
		if log_err == nil {
			break
		}
		time.Sleep(10 * time.Second)
		attempt--
		if attempt == 0 {
			return nil, fmt.Errorf("failed to get the logs in 5 attempts. network id: %s block number %d", c.Network.Id, block_number)
		}
	}
	err = block.SetLogs(b, raw_logs)

	return b, err
}

// Returns the block logs
func (c *NetworkClient) GetBlockLogs(block_number uint64) ([]eth_types.Log, error) {
	big_int := big.NewInt(int64(block_number))

	query := ethereum.FilterQuery{
		FromBlock: big_int,
		ToBlock:   big_int,
		Addresses: []eth_common.Address{},
	}

	raw_logs, log_err := c.client.FilterLogs(c.ctx, query)
	return raw_logs, log_err
}
