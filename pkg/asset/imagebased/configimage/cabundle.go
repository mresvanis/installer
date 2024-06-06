package configimage

import (
	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/tls"
)

// ImageBasedKubeAPIServerCompleteCABundle is the asset the generates the kube-apiserver-complete-server-ca-bundle,
// which contains all the certs that are valid to confirm the kube-apiserver identity and it also contains the
// Ingress Operator CA certificate.
type ImageBasedKubeAPIServerCompleteCABundle struct {
	tls.CertBundle
}

var _ asset.Asset = (*ImageBasedKubeAPIServerCompleteCABundle)(nil)

// Dependencies returns the dependency of the cert bundle.
func (a *ImageBasedKubeAPIServerCompleteCABundle) Dependencies() []asset.Asset {
	return []asset.Asset{
		&tls.KubeAPIServerLocalhostCABundle{},
		&tls.KubeAPIServerServiceNetworkCABundle{},
		&tls.KubeAPIServerLBCABundle{},
		&IngressOperatorCABundle{},
	}
}

// Generate generates the cert bundle based on its dependencies.
func (a *ImageBasedKubeAPIServerCompleteCABundle) Generate(deps asset.Parents) error {
	var certs []tls.CertInterface
	for _, asset := range a.Dependencies() {
		deps.Get(asset)
		certs = append(certs, asset.(tls.CertInterface))
	}
	return a.CertBundle.Generate("kube-apiserver-complete-server-ca-bundle", certs...)
}

// Name returns the human-friendly name of the asset.
func (a *ImageBasedKubeAPIServerCompleteCABundle) Name() string {
	return "Certificate (kube-apiserver-complete-server-ca-bundle)"
}
