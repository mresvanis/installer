package configimage

import (
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	aiv1beta1 "github.com/openshift/assisted-service/api/v1beta1"
	imagebasedAsset "github.com/openshift/installer/pkg/asset/imagebased"
	"github.com/openshift/installer/pkg/asset/installconfig"
	"github.com/openshift/installer/pkg/ipnet"
	"github.com/openshift/installer/pkg/types"
	"github.com/openshift/installer/pkg/types/imagebased"
)

const (
	rawNMStateConfig = `
interfaces:
  - name: eth0
    type: ethernet
    state: up
    mac-address: 00:00:00:00:00:00
    ipv4:
      enabled: true
      address:
        - ip: 192.168.122.2
          prefix-length: 23
      dhcp: false`
)

func optionalInstallConfig() *imagebasedAsset.OptionalInstallConfig {
	_, newCidr, _ := net.ParseCIDR("192.168.111.0/24")
	_, machineNetCidr, _ := net.ParseCIDR("10.10.11.0/24")

	return &imagebasedAsset.OptionalInstallConfig{
		AssetBase: installconfig.AssetBase{
			Config: &types.InstallConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ocp-ibi-cluster-0",
					Namespace: "cluster-0",
				},
				BaseDomain: "testing.com",
				PullSecret: testSecret,
				SSHKey:     testSSHKey,
				ControlPlane: &types.MachinePool{
					Name:     "controlplane",
					Replicas: ptr.To[int64](1),
					Platform: types.MachinePoolPlatform{},
				},
				Networking: &types.Networking{
					MachineNetwork: []types.MachineNetworkEntry{
						{
							CIDR: ipnet.IPNet{IPNet: *machineNetCidr},
						},
					},
					ClusterNetwork: []types.ClusterNetworkEntry{
						{
							CIDR:       ipnet.IPNet{IPNet: *newCidr},
							HostPrefix: 23,
						},
					},
					ServiceNetwork: []ipnet.IPNet{
						*ipnet.MustParseCIDR("172.30.0.0/16"),
					},
					NetworkType: "OVNKubernetes",
				},
				Platform: types.Platform{
					// None: types.Platform.None,
				},
			},
		},
		Supplied: true,
	}
}

func imageBasedConfig() *ImageBasedConfig {
	ibConfig := &ImageBasedConfig{
		Config: &imagebased.Config{
			ObjectMeta: metav1.ObjectMeta{
				Name: "imagebased-config-cluster0",
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: imagebased.ImageBasedConfigVersion,
			},
			Hostname:        "somehostname",
			ReleaseRegistry: "quay.io",
			NetworkConfig: aiv1beta1.NetConfig{
				Raw: unmarshalJSON([]byte(rawNMStateConfig)),
			},
		},
	}
	return ibConfig
}

func unmarshalJSON(b []byte) []byte {
	output, _ := yaml.JSONToYAML(b)
	return output
}
