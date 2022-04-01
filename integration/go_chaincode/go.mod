module github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode

go 1.16

replace (
	github.com/fsouza/go-dockerclient => github.com/fsouza/go-dockerclient v1.4.1
	github.com/go-kit/kit => github.com/go-kit/kit v0.7.0
	github.com/hyperledger-labs/fabric-smart-client => github.com/hyperledger-labs/fabric-smart-client v0.0.0-20220326092546-fb374c9be12b
	github.com/hyperledger/fabric => github.com/hyperledger/fabric v1.4.0-rc1.0.20210722174351-9815a7a8f0f7
	github.com/hyperledger/fabric-private-chaincode => ../..
	github.com/hyperledger/fabric-protos-go => github.com/hyperledger/fabric-protos-go v0.0.0-20201028172056-a3136dde2354
	github.com/libp2p/go-libp2p-core => github.com/libp2p/go-libp2p-core v0.3.0
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20181228115726-23731bf9ba55
)

require (
	github.com/hyperledger-labs/fabric-smart-client v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric v2.1.1+incompatible
	github.com/libp2p/go-libp2p-core v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.1
)
