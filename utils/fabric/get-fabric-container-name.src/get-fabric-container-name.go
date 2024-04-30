/*
* Copyright 2019 Intel Corporation
*
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"regexp"
	"strings"

	"github.com/hyperledger/fabric-lib-go/bccsp"
	"github.com/hyperledger/fabric-lib-go/bccsp/factory"
)

func main() {
	netId := flag.String("net-id", "dev", "peer->networkId as specified in core.yaml")
	peerId := flag.String("peer-id", "jdoe", "peer->Id as specified in core.yaml")
	ccName := flag.String("cc-name", "ecc", "name of CC")
	ccVersion := flag.String("cc-version", "0", "version of CC")

	flag.Parse()

	// chaincode id consists of name and version, see https://github.com/hyperledger/fabric/blob/c491d69962966db1f0231496ae6cab457d8a247d/core/scc/scc.go#L24
	ccid := *ccName + ":" + *ccVersion
	name, _ := getVMNameForDocker(ccid, *netId, *peerId)
	fmt.Println(name)
}

var vmRegExp = regexp.MustCompile("[^a-zA-Z0-9-_.]")

func getVMNameForDocker(ccid, netId, peerId string) (string, error) {

	name := ccid

	if netId != "" && peerId != "" {
		name = fmt.Sprintf("%s-%s-%s", netId, peerId, name)
	} else if netId != "" {
		name = fmt.Sprintf("%s-%s", netId, name)
	} else if peerId != "" {
		name = fmt.Sprintf("%s-%s", peerId, name)
	}

	name = strings.ReplaceAll(name, ":", "-")
	hash := hex.EncodeToString(computeSHA256([]byte(name)))
	saniName := vmRegExp.ReplaceAllString(name, "-")
	imageName := strings.ToLower(fmt.Sprintf("%s-%s", saniName, hash))

	return imageName, nil
}

func computeSHA256(data []byte) (hash []byte) {
	hash, err := factory.GetDefault().Hash(data, &bccsp.SHA256Opts{})
	if err != nil {
		panic(fmt.Errorf("failed computing SHA256 on [% x]", data))
	}
	return
}
