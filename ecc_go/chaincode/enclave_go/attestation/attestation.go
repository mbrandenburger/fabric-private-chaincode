/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/ego_dcap"
	"github.com/hyperledger/fabric-private-chaincode/internal/attestation/simulation"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"
)

func Issue(attestedData *anypb.Any) ([]byte, error) {
	issuer := ego_dcap.NewEgoDcapIssuer()
	att, err := issuer.Issue(attestedData.Value)
	if err != nil {
		if err.Error() == "OE_UNSUPPORTED" {
			logger.Warnf("HW mode attestation failed - continue with simulation mode attestation")
			// When running with SGX Simulation mode; Issue returns OE_UNSUPPORTED;
			// https://github.com/edgelesssys/marblerun/blob/590f82fcd4bd6885ecdd22037b243c82fdc992a1/coordinator/core/core.go#L331
			// In that case we use the simulation attestation as fallback
			issuer := simulation.NewSimulationIssuer()
			return issuer.Issue(attestedData.Value)
		}
		return nil, errors.Wrap(err, "cannot get attestation")
	}

	return att, nil
}
