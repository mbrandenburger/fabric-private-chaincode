/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/gateway-api/proposal"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var now = time.Now()
var assetId = fmt.Sprintf("asset%d", now.Unix()*1e3+int64(now.Nanosecond())/1e6)

func main() {

	peerEndpoint := os.Getenv("CORE_PEER_ADDRESS")
	gatewayPeer := os.Getenv("CORE_PEER_ID")
	mspPath := os.Getenv("CORE_PEER_MSPCONFIGPATH")
	mspID := os.Getenv("CORE_PEER_LOCALMSPID")

	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection(peerEndpoint, gatewayPeer)
	defer clientConnection.Close()

	certPath, err := findSigningCert(mspPath)
	if err != nil {
		panic(err)
	}

	id := newIdentity(mspID, certPath)
	signer := newSigner(path.Join(mspPath, "keystore"))

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(signer.Sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "basic"
	if ccname := os.Getenv("CC_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)
	//contract := network.GetContract(chaincodeName)
	ercc := network.GetContract("ercc")

	// fetch enclave endpoints from ercc
	endpoints := fetchFPCEndpoints(ercc, chaincodeName)

	// fetch enclave encryption key from ercc
	//ep := &crypto.EncryptionProviderImpl{
	//	CSP: crypto.GetDefaultCSP(),
	//	GetCcEncryptionKey: func() ([]byte, error) {
	//		// Note that this function is called during EncryptionProvider.NewEncryptionContext()
	//		return fetchFPCCCKey(ercc, chaincodeName), nil
	//	}}

	// establish grpc connection to endpoint
	if len(endpoints) != 1 {
		panic("we need assume just a single endpoint at this point!")
	}

	fmt.Printf("fpc endpoint: %v\n", endpoints)

	con := newGrpcConnection(endpoints[0], gatewayPeer)
	fpcPeer := peer.NewEndorserClient(con)

	//fpcctx, err := ep.NewEncryptionContext()
	//if err != nil {
	//	panic(err)
	//}

	//encryptedRequest, err := fpcctx.Conceal("hello", []string{"world"})
	//if err != nil {
	//	panic(err)
	//}

	fmt.Printf("id: %v\n", id)

	//encryptedRequest := base64.StdEncoding.EncodeToString([]byte("hello"))

	proposalProto, err := proposal.NewProposal(id, chaincodeName, "__invoke", proposal.WithChannel(channelName), proposal.WithArguments([]byte("hello")))
	if err != nil {
		panic(err)
	}

	signedProposal, err := proposal.NewSignedProposal(proposalProto, signer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("signed proposal: %v\n", signedProposal)

	//ecdsa.VerifyASN1()

	// send signed proposal
	proposalResponse, err := fpcPeer.ProcessProposal(context.Background(), signedProposal)
	if err != nil {
		panic(fmt.Errorf("failed to __invoke FPC chaincode: %w", err))
	}

	fmt.Printf("resp: %v\n", proposalResponse)

}

func fetchFPCEndpoints(contract *client.Contract, chaincodeName string) []string {
	evaluateResult, err := contract.EvaluateTransaction("queryChaincodeEndPoints", chaincodeName)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	return strings.Split(string(evaluateResult), ",")
}

func fetchFPCCCKey(contract *client.Contract, chaincodeName string) []byte {
	evaluateResult, err := contract.EvaluateTransaction("queryChaincodeEncryptionKey", chaincodeName)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	return evaluateResult
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection(peerEndpoint, gatewayPeer string) *grpc.ClientConn {
	// TODO add certs for the endpoint friend

	tlsCertPath := os.Getenv("CORE_PEER_TLS_ROOTCERT_FILE")

	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity(mspID, certPath string) *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

func findSigningCert(mspConfigPath string) (string, error) {
	p := filepath.Join(mspConfigPath, "signcerts")
	files, err := os.ReadDir(p)
	if err != nil {
		return "", errors.Wrapf(err, "error while searching pem in %s", mspConfigPath)
	}

	// return first pem we find
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pem") {
			return filepath.Join(p, f.Name()), nil
		}
	}

	return "", errors.Errorf("cannot find pem in %s", mspConfigPath)
}

type signer struct {
	f identity.Sign
}

func (s *signer) Sign(in []byte) ([]byte, error) {
	return s.f(in)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSigner(keyPath string) *signer {
	files, err := os.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := os.ReadFile(path.Join(keyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return &signer{f: sign}
}
