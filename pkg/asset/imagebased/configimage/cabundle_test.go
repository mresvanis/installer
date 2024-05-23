package configimage

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/tls"
)

func TestCaBundle_Generate(t *testing.T) {
	expectedBundleRaw := bytes.Join([][]byte{
		lbCABundle().BundleRaw,
		localhostCABundle().BundleRaw,
		serviceNetworkCABundle().BundleRaw,
		ingressCABundle().BundleRaw,
	}, []byte{})

	cases := []struct {
		name         string
		dependencies []asset.Asset
		expected     *tls.CertBundle
	}{
		{
			name: "valid dependencies",
			dependencies: []asset.Asset{
				lbCABundle(),
				localhostCABundle(),
				serviceNetworkCABundle(),
				ingressCABundle(),
			},
			expected: &tls.CertBundle{
				BundleRaw: expectedBundleRaw,
				FileList: []*asset.File{
					{
						Filename: "tls/kube-apiserver-complete-server-ca-bundle.crt",
						Data:     expectedBundleRaw,
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parents := asset.Parents{}
			parents.Add(tc.dependencies...)

			asset := &ImageBasedKubeAPIServerCompleteCABundle{}
			err := asset.Generate(parents)
			assert.NoError(t, err)
			assert.Equal(t, string(tc.expected.BundleRaw), string(asset.CertBundle.BundleRaw))
			assert.Equal(t, tc.expected.FileList, asset.CertBundle.FileList)
		})
	}
}

const (
	testCert = `-----BEGIN CERTIFICATE-----
MIICYTCCAcqgAwIBAgIJAI2kA+uXAbhOMA0GCSqGSIb3DQEBCwUAMEgxCzAJBgNV
BAYTAlVTMQswCQYDVQQIDAJDQTEWMBQGA1UEBwwNU2FuIEZyYW5jaXNjbzEUMBIG
A1UECgwLUmVkIEhhdCBJbmMwHhcNMTkwMjEyMTkzMjUzWhcNMTkwMjEzMTkzMjUz
WjBIMQswCQYDVQQGEwJVUzELMAkGA1UECAwCQ0ExFjAUBgNVBAcMDVNhbiBGcmFu
Y2lzY28xFDASBgNVBAoMC1JlZCBIYXQgSW5jMIGfMA0GCSqGSIb3DQEBAQUAA4GN
ADCBiQKBgQC+HOC0mKig/oINAKPo88LqxDJ4l7lozdLtp5oGeqWrLUXSfkvXAkQY
2QYdvPAjpRfH7Ii7G0Asx+HTKdvula7B5fXDjc6NYKuEpTJZRV1ugntI97bozF/E
C2BBmxxEnJN3+Xe8RYXMjz5Q4aqPw9vZhlWN+0hrREl1Ea/zHuWFIQIDAQABo1Mw
UTAdBgNVHQ4EFgQUvTS1XjlvOdsufSyWxukyQu3LriEwHwYDVR0jBBgwFoAUvTS1
XjlvOdsufSyWxukyQu3LriEwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsF
AAOBgQB9gFcOXnzJrM65QqxeCB9Z5l5JMjp45UFC9Bj2cgwDHP80Zvi4omlaacC6
aavmnLd67zm9PbYDWRaOIWAMeB916Iwaw/v6I0jwhAk/VxX5Fl6cGlZu9jZ3zbFE
2sDqkwzIuSjCG2A23s6d4M1S3IXCCydoCSLMu+WhLkbboK6jEg==
-----END CERTIFICATE-----
`
)

func lbCABundle() *tls.KubeAPIServerLBCABundle {
	return &tls.KubeAPIServerLBCABundle{
		CertBundle: tls.CertBundle{
			BundleRaw: []byte(testCert),
		},
	}
}

func localhostCABundle() *tls.KubeAPIServerLocalhostCABundle {
	return &tls.KubeAPIServerLocalhostCABundle{
		CertBundle: tls.CertBundle{
			BundleRaw: []byte(testCert),
		},
	}
}

func serviceNetworkCABundle() *tls.KubeAPIServerServiceNetworkCABundle {
	return &tls.KubeAPIServerServiceNetworkCABundle{
		CertBundle: tls.CertBundle{
			BundleRaw: []byte(testCert),
		},
	}
}

func ingressCABundle() *IngressOperatorCABundle {
	return &IngressOperatorCABundle{
		CertBundle: tls.CertBundle{
			BundleRaw: []byte(testCert),
		},
	}
}
