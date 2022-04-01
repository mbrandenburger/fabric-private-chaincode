//go:build WITH_EGO_DCAP

package attestation

func init() {
	registry.add(ego_dcap.NewEgoDcapVerifier())
}
