// Copyright 2022 PolyCrypt GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"math/big"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wire"
	"perun.network/perun-examples/app-channel/app"
	"perun.network/perun-examples/app-channel/client"
)

const (
	chainURL = "ws://127.0.0.1:8545"
	chainID  = 1337

	// Private keys.
	keyDeployer = "79ea8f62d97bc0591a4224c1725fca6b00de5b2cea286fe2e0bb35c5e76be46e"
	keyServer    = "1af2e950272dd403de7a5760d41c6e44d92b6d02797e51810795ff03cc2cda4f"
	keyClient      = "f63d7d8e930bccd74e93cf5662fde2c28fd8be95edb70c73f1bdd863d07f412e"
)

// main runs a demo of the game client. It assumes that a blockchain node is
// available at `chainURL` and that the accounts corresponding to the specified
// secret keys are provided with sufficient funds.
func main() {
	// Deploy contracts.
	log.Println("Deploying contracts.")
	adjudicator, assetHolder, appAddress := deployContracts(chainURL, chainID, keyDeployer)
	asset := *ethwallet.AsWalletAddr(assetHolder)
	app := app.NewFLApp(ethwallet.AsWalletAddr(appAddress))

	// Setup clients.
	log.Println("Setting up clients.")
	bus := wire.NewLocalBus() // Message bus used for off-chain communication.
	stake := client.EthToWei(big.NewFloat(5))
	server := setupGameClient(bus, chainURL, adjudicator, asset, keyServer, app, stake)
	client := setupGameClient(bus, chainURL, adjudicator, asset, keyClient, app, stake)

	// Print balances before transactions.
	l := newBalanceLogger(chainURL)
	l.LogBalances(server, client)

	// Open app channel and play.
	log.Println("Opening channel.")
	appServer := server.OpenAppChannel(client.WireAddress())
	appClient := client.AcceptedChannel()


	// Set(model, numberOfRounds, weight, accuracy, loss int, actorIdx channel.Index)
	log.Println("Start playing.")
	// round 1

	model := "qmqfkfa74gj4t1xwqfx25hbpt7zmuye3mbpoxru1hzyuas"
	aggregate1 := "amqfkfa74gj4t1xwqfx25hbpt7zmuye3mbpoxru1hzyuas"
	aggregate2 := "bmqfkfa74gj4t1xwqfx25hbpt7zmuye3mbpoxru1hzyuas"
	aggregate3 := "cmqfkfa74gj4t1xwqfx25hbpt7zmuye3mbpoxru1hzyuas"

	weight1 := "1mqfkfa74gj4t1xwqfx25hbpt7zmuye3mbpoxru1hzyuas"
	weight2 := "2mQFKFA74GJ4t1xwQFx25HbpT7ZmuyE3MBpoxRu1hZyuaS"
	weight3 := "3mQFKFA74GJ4t1xwQFx25HbpT7ZmuyE3MBpoxRu1hZyuaS"


	appServer.ForceSet(model, 3, model, 0, 0)

	log.Println("Client's turn.")
	appClient.ForceSet(model, 3, weight1, 0, 0)

	log.Println("Server's turn.")
	appServer.ForceSet(aggregate1, 3, weight1, 66, 44)

	// round 2
	log.Println("Client's turn.")
	appClient.ForceSet(model, 3, weight2, 66, 44)

	log.Println("Server's turn.")
	appServer.ForceSet(aggregate2, 3, weight2, 67, 43)

	// round 3
	log.Println("Client's turn.")
	appClient.ForceSet(model, 3, weight3, 67, 43)

	log.Println("Server's turn.")
	appServer.ForceSet(aggregate3, 3, weight3, 68, 42)

	// // Dispute channel state.
	// log.Println("Server's turn.")
	// appServer.ForceSet(1, 2)

	log.Println("Client wins.")
	log.Println("Payout.")

	// Payout.
	appServer.Settle()
	appClient.Settle()

	// Print balances after transactions.
	l.LogBalances(server, client)

	// Cleanup.
	server.Shutdown()
	client.Shutdown()
}
