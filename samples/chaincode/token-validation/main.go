package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/hyperledger-labs/fabric-token-sdk/token"
	"github.com/hyperledger-labs/fabric-token-sdk/token/services/tcc"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	fpc "github.com/hyperledger/fabric-private-chaincode/ecc_go/chaincode"
)

type serverConfig struct {
	CCID           string
	CCaddress      string
	LogLevel       string
	MetricsEnabled bool
	MetricsServer  string
}

func main() {
	metricsEnabledEnv := os.Getenv("CHAINCODE_METRICS_ENABLED")
	metricsEnabled := false
	if len(metricsEnabledEnv) > 0 {
		var err error
		metricsEnabled, err = strconv.ParseBool(metricsEnabledEnv)
		if err != nil {
			fmt.Printf("Error parsing CHAINCODE_METRICS_ENABLED: %s\n", err)
			os.Exit(1)
		}
	}

	config := serverConfig{
		CCID:           os.Getenv("CHAINCODE_PKG_ID"),
		CCaddress:      os.Getenv("CHAINCODE_SERVER_ADDRESS"),
		LogLevel:       os.Getenv("CHAINCODE_LOG_LEVEL"),
		MetricsEnabled: metricsEnabled,
		MetricsServer:  os.Getenv("CHAINCODE_METRICS_SERVER"),
	}
	if len(config.MetricsServer) == 0 {
		config.MetricsServer = "localhost:8125"
	}

	fmt.Printf("metrics server at [%s], enabled [%v]\n", config.MetricsServer, config.MetricsEnabled)

	// create private chaincode
	privateChaincode := fpc.NewPrivateChaincode(&tcc.TokenChaincode{
		TokenServicesFactory: func(bytes []byte) (tcc.PublicParametersManager, tcc.Validator, error) {
			return token.NewServicesFromPublicParams(bytes)
		},
		// inject custom token extractor - since FPC does not yet support transient data
		ExtractTokenRequest: extractTokenRequest,
		LogLevel:            config.LogLevel,
		MetricsEnabled:      config.MetricsEnabled,
		MetricsServer:       config.MetricsServer,
	})

	fmt.Println("starting FPC-TCC... ")

	// start chaincode as a service
	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.CCaddress,
		CC:      privateChaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true, // just for testing good enough
		},
	}

	if err := server.Start(); err != nil {
		panic(err)
	}
}

func extractTokenRequest(stub shim.ChaincodeStubInterface) ([]byte, error) {
	args := stub.GetArgs()
	if len(args) != 2 {
		return nil, errors.New("empty token request")
	}
	// extract token request from transient
	return args[1], nil
}
