// Package handler defines the commands and the command handlers
// exposed by categorizer service's reply controller and pull controller.
package handler

import (
	"github.com/blocklords/sds/app/command"
)

const (
	// Get the list of decoded smartcontracts by the
	// [github.com/blocklords/sds/common/topic.TopicFilter]
	//
	// Users are calling requesting this command through SDK.
	// Intended to be access through the router
	SNAPSHOT command.CommandName = "snapshot_get"
	// Get all smartcontracts and the categorization state from this service
	// Intended to be called directly
	GET_SMARTCONTRACTS command.CommandName = "smartcontract_get_all"
	// Get the smartcontract and it's categorization state from this service
	// Indended to be called directly
	GET_SMARTCONTRACT command.CommandName = "smartcontract_get"
	// Add a new smartcontract to categorize.
	//
	// This service then will call blockchain's sub categorizer services
	// Through the router
	SET_SMARTCONTRACT command.CommandName = "smartcontract_set"
	// CATEGORIZATION command is sent from blockchain sub services to this service
	// with the list of decoded smartcontract logs and new states.
	//
	// Internal from SDS network services to SDS Categorizer
	// Indicates that the list of smartcontracts are categorized
	CATEGORIZATION command.CommandName = "categorize"
)

// Return the list of command handlers for this service
// For the controller
func CommandHandlers() command.Handlers {
	return command.EmptyHandlers().
		Add(GET_SMARTCONTRACTS, GetSmartcontracts).
		Add(GET_SMARTCONTRACT, GetSmartcontract).
		Add(SET_SMARTCONTRACT, SetSmartcontract).
		Add(SNAPSHOT, GetSnapshot).
		Add(CATEGORIZATION, on_new_smartcontracts)
}
