package client

import (
	"context"
	"time"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
)

func New(host string) (*ccClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	chaincodeClient := pb.NewChaincodeClient(conn)
	stream, err := chaincodeClient.Connect(context.Background())
	return &ccClient{
		conn:   conn,
		stream: stream,
	}, nil
}

type ccClient struct {
	conn   *grpc.ClientConn
	stream pb.Chaincode_ConnectClient
}

func (c *ccClient) Close() {
	c.conn.Close()
}

func (c *ccClient) SendMsg(msg *pb.ChaincodeMessage) error {
	return c.stream.Send(msg)
}

func (c *ccClient) Recv() (*pb.ChaincodeMessage, error) {
	return c.stream.Recv()
}
