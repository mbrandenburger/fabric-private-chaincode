/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ego_dcap

import "github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"

const EgoDcapType = "ego-dcap"

// NewEgoDcapConverter creates a new attestation converter for Intel SGX simulation mode
func NewEgoDcapConverter() *types.Converter {
	return &types.Converter{
		Type: EgoDcapType,
		Converter: func(attestationBytes []byte) (evidenceBytes []byte, err error) {
			// NO-OP
			return attestationBytes, nil
		},
	}
}
