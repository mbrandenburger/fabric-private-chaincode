package main

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-private-chaincode/utils/fabric/mock-peer/pkg/client"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

func main() {
	cc, err := client.New("localhost:8087")
	must(err)
	defer cc.Close()

	registeredMsg := &pb.ChaincodeMessage{
		Type: pb.ChaincodeMessage_REGISTERED,
	}

	readyMsg := &pb.ChaincodeMessage{
		Type: pb.ChaincodeMessage_READY,
	}

	initEnclaveMsg := createInvcation([][]byte{[]byte("__initEnclave"), []byte("arg1"), []byte("arg2")})

	// receive register request
	resp, err := cc.Recv()
	must(err)
	fmt.Printf("resp: %v\n", resp)

	// send ack register
	err = cc.SendMsg(registeredMsg)
	must(err)

	// send ready
	err = cc.SendMsg(readyMsg)
	must(err)

	// send tx invocation
	err = cc.SendMsg(initEnclaveMsg)
	must(err)

	// receive message
	resp, err = cc.Recv()
	must(err)
	fmt.Printf("resp: %v\n", resp)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func createInvcation(args [][]byte) *pb.ChaincodeMessage {

	payload := &pb.ChaincodeProposalPayload{}
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
		Signature:     []byte("mock_signature"),
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
		Txid:      "fake_tx_id",
		ChannelId: "fake_channel",
		Payload:   inputBytes,
		Proposal:  signedProposal,
	}

	return invokeMsg
}
