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
	// "log"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"time"

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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}

// main runs a demo of the game client. It assumes that a blockchain node is
// available at `chainURL` and that the accounts corresponding to the specified
// secret keys are provided with sufficient funds.
func main() {
	// Deploy contracts.
	// log.Println("Deploying contracts.")
	adjudicator, assetHolder, appAddress := deployContracts(chainURL, chainID, keyDeployer)
	asset := *ethwallet.AsWalletAddr(assetHolder)
	app := app.NewFLApp(ethwallet.AsWalletAddr(appAddress))

	// // Setup clients.
	// // log.Println("Setting up clients.")
	bus := wire.NewLocalBus() // Message bus used for off-chain communication.
	stake := client.EthToWei(big.NewFloat(20))
	server := setupGameClient(bus, chainURL, adjudicator, asset, keyServer, app, stake)
	client := setupGameClient(bus, chainURL, adjudicator, asset, keyClient, app, stake)

	// Print balances before transactions.
	l := newBalanceLogger(chainURL)
	l.LogBalances(server, client)

	// // Open app channel and play.
	// log.Println("Opening channel.")

	idDisputed := false

	elapsedTimeMapChannelOpening := make(map[int][]string)
	elapsedTimeMapChannelClosing := make(map[int][]string)
	elapsedTimeMap := make(map[int][]string)

	// repeat 10 times
	repeat := 10


	for rep := 2; rep <= repeat; rep++ {

		elapsedTimeMap[rep] = make([]string, 10)
		elapsedTimeMapChannelOpening[rep] = make([]string, 10)
		elapsedTimeMapChannelClosing[rep] = make([]string, 10)
		// do experiment for 1 to 10 rounds
		for i := 1; i <= 10; i++ {
			startTimeChannelOpening := time.Now()

			appServer := server.OpenAppChannel(client.WireAddress())
			appClient := client.AcceptedChannel()

			elapsedTimeMapChannelOpening[rep][i-1] = time.Since(startTimeChannelOpening).String()

			startTime := time.Now()
			numberOfRounds := i
			model := RandStringBytes(46)

			appServer.Set(model, numberOfRounds, model, 0, 0)

				for r := 1; r <= numberOfRounds; r++ {
					weight := RandStringBytes(46)
					aggregate := RandStringBytes(46)

					if idDisputed {
						appClient.ForceSet(model, numberOfRounds, weight, 0, 0)
					} else {
						appClient.Set(model, numberOfRounds, weight, 0, 0)
					}
					if idDisputed {
						appServer.ForceSet(aggregate, numberOfRounds, weight, 67, 43)
					} else {
						appServer.Set(aggregate, numberOfRounds, weight, 67, 43)
					}

					model = aggregate

				}

			elapsedTime := time.Since(startTime)
			// // convert to milliseconds
			// store elapsed time into a map
			elapsedTimeMap[rep][i-1] = elapsedTime.String()

			startTimeChannelClosing := time.Now()
			appServer.Settle()
			appClient.Settle()
			elapsedTimeMapChannelClosing[rep][i-1] = time.Since(startTimeChannelClosing).String()
		}
		j, err := json.Marshal(elapsedTimeMap)
		fmt.Println(string(j), err)
		fileName := fmt.Sprintf("disputes_%v.json", rep)
		_ = os.WriteFile(fileName, j, 0644)

		j, err = json.Marshal(elapsedTimeMapChannelOpening)
		fmt.Println(string(j), err)
		fileName = fmt.Sprintf("opening_channel%v.json", rep)
		_ = os.WriteFile(fileName, j, 0644)

		j, err = json.Marshal(elapsedTimeMapChannelClosing)
		fmt.Println(string(j), err)
		fileName = fmt.Sprintf("closing_channel%v.json", rep)
		_ = os.WriteFile(fileName, j, 0644)

	}

	// Cleanup.
	server.Shutdown()
	client.Shutdown()
}
