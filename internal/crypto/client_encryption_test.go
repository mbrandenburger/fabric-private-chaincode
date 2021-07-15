package crypto

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
	"github.com/hyperledger/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/test-go/testify/assert"
)

func TestNewEncryptionContext(t *testing.T) {
	provider := &EncryptionProviderImpl{
		func() ([]byte, error) {
			return nil, fmt.Errorf("some error while fetching key")
		},
	}
	expectedErrorMsg := "failed to get chaincode encryption key from ercc: some error while fetching key"
	ctx, err := provider.NewEncryptionContext()
	assert.Nil(t, ctx)
	assert.Error(t, err, expectedErrorMsg)

	provider = &EncryptionProviderImpl{
		func() ([]byte, error) {
			return []byte("some invalid base64 encoded key"), nil
		},
	}
	ctx, err = provider.NewEncryptionContext()
	assert.Nil(t, ctx)
	assert.Error(t, err)

	provider = &EncryptionProviderImpl{
		func() ([]byte, error) {
			return []byte(base64.StdEncoding.EncodeToString([]byte("some key"))), nil
		},
	}
	ctx, err = provider.NewEncryptionContext()
	assert.NotNil(t, ctx)
	assert.NoError(t, err)
}

func TestConceal(t *testing.T) {
	f := "some function"
	args := []string{"some", "args"}

	// test with some invalid request encryption key
	ctxImpl := &EncryptionContextImpl{
		requestEncryptionKey: []byte("invalid request encryption key"),
	}
	request, err := ctxImpl.Conceal(f, args)
	assert.Empty(t, request)
	assert.Error(t, err)

	// test with some invalid request encryption key
	symKey, err := NewSymmetricKey()
	assert.NoError(t, err)
	ctxImpl = &EncryptionContextImpl{
		requestEncryptionKey:   symKey,
		chaincodeEncryptionKey: []byte("invalid chaincode encryption key"),
	}
	request, err = ctxImpl.Conceal(f, args)
	assert.Empty(t, request)
	assert.Error(t, err)

	// test with valid rsa key
	pubKey, privKey, err := NewRSAKeys()
	assert.NotNil(t, pubKey)
	assert.NotNil(t, privKey)
	assert.NoError(t, err)
	provider := &EncryptionProviderImpl{
		func() ([]byte, error) {
			return []byte(base64.StdEncoding.EncodeToString(pubKey)), nil
		},
	}
	ctx, err := provider.NewEncryptionContext()
	assert.NotNil(t, ctx)
	assert.NoError(t, err)

	// should succeed
	request, err = ctx.Conceal(f, args)
	assert.NotEmpty(t, request)
	assert.NoError(t, err)

	_, err = base64.StdEncoding.DecodeString(request)
	assert.NoError(t, err)
}

func TestReveal(t *testing.T) {
	msg := []byte("some response")

	pubKey, privKey, err := NewRSAKeys()
	assert.NotNil(t, pubKey)
	assert.NotNil(t, privKey)
	assert.NoError(t, err)

	requestEncryptionKey, err := NewSymmetricKey()
	assert.NoError(t, err)

	responseEncryptionKey, err := NewSymmetricKey()
	assert.NoError(t, err)

	ctx := &EncryptionContextImpl{
		requestEncryptionKey:   requestEncryptionKey,
		responseEncryptionKey:  responseEncryptionKey,
		chaincodeEncryptionKey: pubKey,
	}

	// test different invalid inputs
	resp, err := ctx.Reveal(nil)
	assert.Nil(t, resp)
	assert.Error(t, err)

	resp, err = ctx.Reveal([]byte("invalid input (not base64)"))
	assert.Nil(t, resp)
	assert.Error(t, err)

	resp, err = ctx.Reveal([]byte(base64.StdEncoding.EncodeToString([]byte("not a SignedChaincodeResponseMessage"))))
	assert.Nil(t, resp)
	assert.Error(t, err)

	resp, err = ctx.Reveal([]byte(utils.MarshallProto(&protos.SignedChaincodeResponseMessage{})))
	assert.Nil(t, resp)
	assert.Error(t, err)

	resp, err = ctx.Reveal([]byte(utils.MarshallProto(&protos.SignedChaincodeResponseMessage{ChaincodeResponseMessage: []byte("some invalid response")})))
	assert.Nil(t, resp)
	assert.Error(t, err)

	// msg not encrypted
	response := &protos.ChaincodeResponseMessage{EncryptedResponse: msg}
	responseBytes := protoutil.MarshalOrPanic(response)
	resp, err = ctx.Reveal([]byte(utils.MarshallProto(&protos.SignedChaincodeResponseMessage{ChaincodeResponseMessage: responseBytes})))
	assert.Nil(t, resp)
	assert.Error(t, err)

	// msg not base64 encoded
	encryptedMsg, err := EncryptMessage(responseEncryptionKey, msg)
	response = &protos.ChaincodeResponseMessage{EncryptedResponse: encryptedMsg}
	responseBytes = protoutil.MarshalOrPanic(response)
	resp, err = ctx.Reveal([]byte(utils.MarshallProto(&protos.SignedChaincodeResponseMessage{ChaincodeResponseMessage: responseBytes})))
	assert.Nil(t, resp)
	assert.Error(t, err)

	// should succeed
	encryptedMsg, err = EncryptMessage(responseEncryptionKey, []byte(base64.StdEncoding.EncodeToString(msg)))
	response = &protos.ChaincodeResponseMessage{EncryptedResponse: encryptedMsg}
	responseBytes = protoutil.MarshalOrPanic(response)
	resp, err = ctx.Reveal([]byte(utils.MarshallProto(&protos.SignedChaincodeResponseMessage{ChaincodeResponseMessage: responseBytes})))
	assert.Equal(t, resp, msg)
	assert.NoError(t, err)
}