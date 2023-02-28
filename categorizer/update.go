package categorizer

import (
	debug_log "log"

	"github.com/blocklords/gosds/app/remote/message"
	"github.com/blocklords/gosds/categorizer/log"
	"github.com/blocklords/gosds/categorizer/smartcontract"
	"github.com/blocklords/gosds/db"

	zmq "github.com/pebbe/zmq4"
)

// Sets up the socket that will be connected by the blockchain/categorizers
// The blockchain categorizers will set up the smartcontract informations on the database
func SetupSocket(database *db.Database) {
	sock, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		panic(err)
	}

	url := "cat"
	if err := sock.Bind("inproc://" + url); err != nil {
		debug_log.Fatalf("trying to create categorizer socket: %v", err)
	}

	for {
		// Wait for reply.
		msgs, _ := sock.RecvMessage(0)
		request, _ := message.ParseRequest(msgs)

		raw_smartcontracts, _ := request.Parameters.GetKeyValueList("smartcontracts")
		smartcontracts := make([]*smartcontract.Smartcontract, len(raw_smartcontracts))

		for i, raw := range raw_smartcontracts {
			sm, _ := smartcontract.New(raw)
			smartcontracts[i] = sm
		}

		raw_logs, _ := request.Parameters.GetKeyValueList("logs")

		logs := make([]*log.Log, len(raw_logs))
		for i, raw := range raw_logs {
			log, _ := log.NewFromMap(raw)
			logs[i] = log
		}

		for _, sm := range smartcontracts {
			err := smartcontract.SetSyncing(database, sm, sm.CategorizedBlockNumber, sm.CategorizedBlockTimestamp)
			if err != nil {
				panic(err)
			}
		}

		for _, l := range logs {
			err := log.Save(database, l)
			if err != nil {
				panic(err)
			}
		}
	}
}