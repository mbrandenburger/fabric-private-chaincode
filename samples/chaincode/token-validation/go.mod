module github.com/hyperledger/fabric-private-chaincode/samples/chaincode/token-validation

go 1.16

replace (
	github.com/hyperledger-labs/fabric-smart-client => github.com/mbrandenburger/fabric-smart-client v0.0.0-20220326071315-0df16e0f03b0
	github.com/hyperledger-labs/fabric-token-sdk => github.com/mbrandenburger/fabric-token-sdk v0.0.0-20220330083744-915e7588077b
	github.com/hyperledger/fabric-private-chaincode => ../../..
)

require (
	github.com/hyperledger-labs/fabric-token-sdk v0.0.0-20220325091407-d2938831f9b5
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20210718160520-38d29fabecb9
	github.com/hyperledger/fabric-private-chaincode v1.0.0-rc3
)
