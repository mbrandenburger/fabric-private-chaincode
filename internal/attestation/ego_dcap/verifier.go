//go:build WITH_EGO
// +build WITH_EGO

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ego_dcap

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/edgelesssys/ego/eclient"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/types"
	"github.com/pkg/errors"
)

func NewEgoDcapVerifier() *types.Verifier {
	return &types.Verifier{
		Type:   EgoDcapType,
		Verify: verify,
	}
}

func verify(evidence *types.Evidence, expectedValidationValues *types.ValidationValues) (err error) {

	reportBytes, err := base64.StdEncoding.DecodeString(evidence.Data)
	if err != nil {
		return errors.Wrap(err, "cannot decode evidence data with base64")
	}

	report, err := eclient.VerifyRemoteReport(reportBytes)
	if err != nil {
		return errors.Wrap(err, "report verification failed")
	}

	hash := sha256.Sum256(expectedValidationValues.Statement)
	if !bytes.Equal(report.Data[:len(hash)], hash[:]) {
		return errors.New("report data does not match the expected validation values")
	}

	expectedMrenclave, err := hex.DecodeString(expectedValidationValues.Mrenclave)
	if err != nil {
		return errors.Wrap(err, "cannot decode expected mrenclave")
	}

	if !bytes.Equal(report.UniqueID, expectedMrenclave) {
		return errors.New("report mrenclave does not match the expected mrenclave")
	}

	return nil
}
