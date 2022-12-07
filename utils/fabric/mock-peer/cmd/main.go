package main

import (
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/sgx"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-private-chaincode/utils/fabric/mock-peer/pkg/client"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

func main() {

	enclaveEndpoint := os.Getenv("CHAINCODE_SERVER_ADDRESS")
	if enclaveEndpoint == "" {
		enclaveEndpoint = "localhost:8087"
	}

	fmt.Printf("Establish connection with %s ...\n", enclaveEndpoint)
	cc, err := client.New(enclaveEndpoint)
	must(err)
	defer cc.Close()

	fmt.Printf("Connected!\n")

	registeredMsg := &pb.ChaincodeMessage{
		Type: pb.ChaincodeMessage_REGISTERED,
	}

	readyMsg := &pb.ChaincodeMessage{
		Type: pb.ChaincodeMessage_READY,
	}

	initEnclaveMsg, err := createInitEnclaveMsg(enclaveEndpoint)
	must(err)

	// receive register request
	fmt.Printf("Initiate handshake protocol ...\n")
	resp, err := cc.Recv()
	must(err)

	fmt.Printf("Received msg: [%v]\n", resp)

	// send ack register
	fmt.Printf("Send register ack\n")
	err = cc.SendMsg(registeredMsg)
	must(err)

	// send ready
	fmt.Printf("Send ready\n")
	err = cc.SendMsg(readyMsg)
	must(err)

	fmt.Printf("Handshake completed!\n")

	// send tx invocation
	fmt.Printf("Send initEnclave invocation\n")
	err = cc.SendMsg(initEnclaveMsg)
	must(err)

	// receive message
	resp, err = cc.Recv()
	must(err)
	fmt.Printf("Received msg: [%v]\n", resp)

	fmt.Printf("Done! Good Bye!\n")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func createInitEnclaveMsg(targetPeer string) (*pb.ChaincodeMessage, error) {

	attestationParams, err := sgx.CreateAttestationParamsFromEnvironment()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get cc params from env")
	}

	serializedJSONParams, err := attestationParams.ToBase64EncodedJSON()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize attestation parameters")
	}
	fmt.Printf("using attestation params: '%v'\n", attestationParams)

	initMsg := &protos.InitEnclaveMessage{
		PeerEndpoint:      targetPeer,
		AttestationParams: serializedJSONParams,
	}

	return createInvcation([][]byte{[]byte("__initEnclave"), []byte(utils.MarshallProtoBase64(initMsg))}), nil
}

func createInvcation(args [][]byte) *pb.ChaincodeMessage {

	invocationSpec := &pb.ChaincodeInvocationSpec{
		ChaincodeSpec: &pb.ChaincodeSpec{
			ChaincodeId: &pb.ChaincodeID{
				Name: "DemoChaincode",
			},
		},
	}
	invocationSpecBytes, err := proto.Marshal(invocationSpec)
	must(err)

	payload := &pb.ChaincodeProposalPayload{
		Input: invocationSpecBytes,
	}
	payloadBytes, err := proto.Marshal(payload)
	must(err)

	shdr := &common.SignatureHeader{
		Creator: []byte("alice"),
		Nonce:   []byte("some_nonce"),
	}
	shdrBytes, err := proto.Marshal(shdr)
	must(err)

	chdr := &common.ChannelHeader{
		Type:  int32(common.HeaderType_ENDORSER_TRANSACTION),
		Epoch: 0,
	}
	chdrBytes, err := proto.Marshal(chdr)
	must(err)

	hdr := &common.Header{
		ChannelHeader:   chdrBytes,
		SignatureHeader: shdrBytes,
	}
	hdrBytes, err := proto.Marshal(hdr)
	must(err)

	proposal := &pb.Proposal{
		Header:  hdrBytes,
		Payload: payloadBytes,
	}
	proposalBytes, err := proto.Marshal(proposal)
	must(err)

	signedProposal := &pb.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     []byte("demo_signature"),
	}

	input := &pb.ChaincodeInput{
		Args:        args,
		Decorations: nil,
		IsInit:      false,
	}
	inputBytes, err := proto.Marshal(input)
	must(err)

	invokeMsg := &pb.ChaincodeMessage{
		Type:      pb.ChaincodeMessage_TRANSACTION,
		Txid:      "demo_tx_id",
		ChannelId: "demo_channel",
		Payload:   inputBytes,
		Proposal:  signedProposal,
	}

	return invokeMsg
}
