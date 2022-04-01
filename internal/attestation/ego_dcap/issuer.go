/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ego_dcap

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/edgelesssys/ego/enclave"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"
)

// NewEgoDcapIssuer creates a new attestation issuer for Intel SGX simulation mode
func NewEgoDcapIssuer() *types.Issuer {
	return &types.Issuer{
		Type:  EgoDcapType,
		Issue: issue,
	}
}

func issue(customData []byte) ([]byte, error) {
	hash := sha256.Sum256(customData)
	report, err := enclave.GetRemoteReport(hash[:])
	if err != nil {
		return nil, err
	}

	att := &types.Attestation{
		Type: EgoDcapType,
		Data: base64.StdEncoding.EncodeToString(report),
	}

	return json.Marshal(att)
}
