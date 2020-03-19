/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-protos-go/common"
	commonerrors "github.com/hyperledger/fabric/common/errors"
	validation "github.com/hyperledger/fabric/core/handlers/validation/api"
	. "github.com/hyperledger/fabric/core/handlers/validation/api/policies"
	. "github.com/hyperledger/fabric/core/handlers/validation/api/state"
	defaultvscc "github.com/hyperledger/fabric/core/handlers/validation/builtin"
	"github.com/pkg/errors"
)

func NewPluginFactory() validation.PluginFactory {
	return &ECCValidationFactory{}
}

type ECCValidationFactory struct {
}

func (*ECCValidationFactory) New() validation.Plugin {
	return &ECCValidation{}
}

type ECCValidation struct {
	DefaultTxValidator validation.Plugin
	ECCTxValidator     TransactionValidator
}

//go:generate mockery -dir . -name TransactionValidator -case underscore -output mocks/
type TransactionValidator interface {
	Validate(txData []byte, policy []byte) commonerrors.TxValidationError
}

func (v *ECCValidation) Validate(block *common.Block, namespace string, txPosition int, actionPosition int, contextData ...validation.ContextDatum) error {
	if len(contextData) == 0 {
		logger.Panicf("Expected to receive policy bytes in context data")
	}

	serializedPolicy, isSerializedPolicy := contextData[0].(SerializedPolicy)
	if !isSerializedPolicy {
		logger.Panicf("Expected to receive a serialized policy in the first context data")
	}
	if block == nil || block.Data == nil {
		return errors.New("empty block")
	}
	if txPosition >= len(block.Data.Data) {
		return errors.Errorf("block has only %d transactions, but requested tx at position %d", len(block.Data.Data), txPosition)
	}
	if block.Header == nil {
		return errors.Errorf("no block header")
	}

	// do defalt vscc
	err := v.DefaultTxValidator.Validate(block, namespace, txPosition, actionPosition, contextData...)
	if err != nil {
		logger.Debugf("block %d, namespace: %s, tx %d validation results is: %v", block.Header.Number, namespace, txPosition, err)
		return convertErrorTypeOrPanic(err)
	}

	// do ecc-vscc
	err = v.ECCTxValidator.Validate(block.Data.Data[txPosition], serializedPolicy.Bytes())
	logger.Debugf("block %d, namespace: %s, tx %d validation results is: %v", block.Header.Number, namespace, txPosition, err)
	return convertErrorTypeOrPanic(err)

}

func convertErrorTypeOrPanic(err error) error {
	if err == nil {
		return nil
	}
	if err, isExecutionError := err.(*commonerrors.VSCCExecutionFailureError); isExecutionError {
		return &validation.ExecutionFailureError{
			Reason: err.Error(),
		}
	}
	if err, isEndorsementError := err.(*commonerrors.VSCCEndorsementPolicyError); isEndorsementError {
		return err
	}
	logger.Panicf("Programming error: The error is %v, of type %v but expected to be either ExecutionFailureError or VSCCEndorsementPolicyError", err, reflect.TypeOf(err))
	return &validation.ExecutionFailureError{Reason: fmt.Sprintf("error of type %v returned from VSCC", reflect.TypeOf(err))}
}

func (v *ECCValidation) Init(dependencies ...validation.Dependency) error {
	var sf StateFetcher
	for _, dep := range dependencies {
		if stateFetcher, isStateFetcher := dep.(StateFetcher); isStateFetcher {
			sf = stateFetcher
		}
	}
	if sf == nil {
		return errors.New("ECC-VSCC: stateFetcher not passed in init")
	}

	v.ECCTxValidator = New(sf)

	// use default vscc and our custom ecc vscc
	factory := &defaultvscc.DefaultValidationFactory{}
	v.DefaultTxValidator = factory.New()
	err := v.DefaultTxValidator.Init(dependencies...)
	if err != nil {
		return errors.Errorf("Error while creating default vscc: %s", err)
	}

	return nil
}
