# Quick and Dirty Mock Peer

Sometimes you just need to test your chaincode without installing it on a peer ...

Do the following:

```bash
cd $FPC_PATH/samples/chaincode/kv-test-go/
CHAINCODE_PKG_ID=dummy CHAINCODE_SERVER_ADDRESS=localhost:8087 go run .
```

Now your chaincode is ready and waits for a peer to connect.
This is where our mock peer comes into the game.

```bash
cd $FPC_PATH/utils/fabric/mock-peer
CHAINCODE_SERVER_ADDRESS=localhost:8087 go run cmd/main.go
```

Via the environment variable `CHAINCODE_SERVER_ADDRESS`, you can define the endpoint of the chaincode.

As the mock peer is just half-baked, we expect to see the following response from the mock peer program.
```bash
resp: type:REGISTER payload:"\022\005dummy"
resp: type:COMPLETED payload:"\010\364\003\022Mgetting initEnclave msg failed: invalid attested data message: unexpected EOF" txid:"fake_tx_id" channel_id:"fake_channel"
```