package configimage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	k8sjson "sigs.k8s.io/json"

	lca "github.com/openshift-kni/lifecycle-agent/api/seedreconfig"
	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/imagebased"
	"github.com/openshift/installer/pkg/asset/installconfig"
	"github.com/openshift/installer/pkg/asset/password"
	"github.com/openshift/installer/pkg/asset/tls"
	"github.com/openshift/installer/pkg/types"
)

const (
	defaultChronyConf = `
pool 0.rhel.pool.ntp.org iburst
driftfile /var/lib/chrony/drift
makestep 1.0 3
rtcsync
logdir /var/log/chrony`

	userCABundleConfigMapName = "user-ca-bundle"
)

var (
	clusterConfigurationFilename = filepath.Join(clusterConfigurationDir, "manifest.json")

	_ asset.WritableAsset = (*ClusterConfiguration)(nil)
)

// ClusterConfiguration generates the image-based installer cluster configuration asset.
type ClusterConfiguration struct {
	File   *asset.File
	Config *lca.SeedReconfiguration
}

// Name returns a human friendly name for the asset.
func (*ClusterConfiguration) Name() string {
	return "Image-based installer cluster configuration"
}

// Dependencies returns all of the dependencies directly needed to generate
// the asset.
func (*ClusterConfiguration) Dependencies() []asset.Asset {
	return []asset.Asset{
		&installconfig.ClusterID{},
		&imagebased.OptionalInstallConfig{},
		&tls.KubeAPIServerLBSignerCertKey{},
		&tls.KubeAPIServerLocalhostSignerCertKey{},
		&tls.KubeAPIServerServiceNetworkSignerCertKey{},
		&tls.AdminKubeConfigSignerCertKey{},
		&IngressOperatorSignerCertKey{},
		&password.KubeadminPassword{},
		&ImageBasedConfig{},
	}
}

// Generate generates the Image-based Installer ClusterConfiguration manifest.
func (cc *ClusterConfiguration) Generate(dependencies asset.Parents) error {
	clusterID := &installconfig.ClusterID{}
	installConfig := &imagebased.OptionalInstallConfig{}
	imageBasedConfig := &ImageBasedConfig{}
	serverLBSignerCertKey := &tls.KubeAPIServerLBSignerCertKey{}
	serverLocalhostSignerCertKey := &tls.KubeAPIServerLocalhostSignerCertKey{}
	serverServiceNetworkSignerCertKey := &tls.KubeAPIServerServiceNetworkSignerCertKey{}
	adminKubeConfigSignerCertKey := &tls.AdminKubeConfigSignerCertKey{}
	ingressOperatorSignerCertKey := &IngressOperatorSignerCertKey{}

	dependencies.Get(
		clusterID,
		installConfig,
		imageBasedConfig,
		serverLBSignerCertKey,
		serverLocalhostSignerCertKey,
		serverServiceNetworkSignerCertKey,
		adminKubeConfigSignerCertKey,
		ingressOperatorSignerCertKey,
	)

	pwd := &password.KubeadminPassword{}
	dependencies.Get(pwd)
	pwdHash := string(pwd.PasswordHash)

	if installConfig.Config == nil || imageBasedConfig.Config == nil {
		return cc.finish()
	}

	cc.Config = &lca.SeedReconfiguration{
		APIVersion:            lca.SeedReconfigurationVersion,
		BaseDomain:            installConfig.Config.BaseDomain,
		ClusterID:             clusterID.UUID,
		ClusterName:           installConfig.ClusterName(),
		Hostname:              imageBasedConfig.Config.Hostname,
		InfraID:               clusterID.InfraID,
		KubeadminPasswordHash: pwdHash,
		Proxy:                 (*lca.Proxy)(installConfig.Config.Proxy),
		PullSecret:            installConfig.Config.PullSecret,
		RawNMStateConfig:      imageBasedConfig.Config.NetworkConfig.String(),
		ReleaseRegistry:       imageBasedConfig.Config.ReleaseRegistry,
		SSHKey:                installConfig.Config.SSHKey,
	}

	if len(imageBasedConfig.Config.AdditionalNTPSources) > 0 {
		cc.Config.ChronyConfig = chronyConfWithAdditionalNTPSources(imageBasedConfig.Config.AdditionalNTPSources)
	}

	if installConfig.Config.AdditionalTrustBundle != "" {
		cc.Config.AdditionalTrustBundle = lca.AdditionalTrustBundle{
			UserCaBundle: installConfig.Config.AdditionalTrustBundle,
		}

		if installConfig.Config.AdditionalTrustBundlePolicy == types.PolicyAlways ||
			(installConfig.Config.AdditionalTrustBundlePolicy == types.PolicyProxyOnly && installConfig.Config.Proxy != nil) {
			cc.Config.AdditionalTrustBundle.ProxyConfigmapName = userCABundleConfigMapName
			cc.Config.AdditionalTrustBundle.ProxyConfigmapBundle = installConfig.Config.AdditionalTrustBundle
		}
	}

	cc.Config.KubeconfigCryptoRetention = lca.KubeConfigCryptoRetention{
		KubeAPICrypto: lca.KubeAPICrypto{
			ServingCrypto: lca.ServingCrypto{
				LoadbalancerSignerPrivateKey:   (lca.PEM)(string(serverLBSignerCertKey.Key())),
				LocalhostSignerPrivateKey:      (lca.PEM)(string(serverLocalhostSignerCertKey.Key())),
				ServiceNetworkSignerPrivateKey: (lca.PEM)(string(serverServiceNetworkSignerCertKey.Key())),
			},
			ClientAuthCrypto: lca.ClientAuthCrypto{
				AdminCACertificate: (lca.PEM)(string(adminKubeConfigSignerCertKey.Cert())),
			},
		},
		IngresssCrypto: lca.IngresssCrypto{
			IngressCA: (lca.PEM)(string(ingressOperatorSignerCertKey.Key())),
		},
	}

	// validation for the length of the MachineNetwork is performed in the imagebased.OptionalInstallConfig
	cc.Config.MachineNetwork = installConfig.Config.Networking.MachineNetwork[0].CIDR.String()

	clusterConfigurationData, err := json.Marshal(cc.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal image-based installer ClusterConfiguration: %w", err)
	}

	cc.File = &asset.File{
		Filename: clusterConfigurationFilename,
		Data:     clusterConfigurationData,
	}

	return cc.finish()
}

// Files returns the files generated by the asset.
func (cc *ClusterConfiguration) Files() []*asset.File {
	if cc.File != nil {
		return []*asset.File{cc.File}
	}
	return []*asset.File{}
}

// Load returns ClusterConfiguration asset from the disk.
func (cc *ClusterConfiguration) Load(f asset.FileFetcher) (bool, error) {
	file, err := f.FetchByName(clusterConfigurationFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to load %s file: %w", clusterConfigurationFilename, err)
	}

	config := &lca.SeedReconfiguration{}
	strErrs, err := k8sjson.UnmarshalStrict(file.Data, config)
	if len(strErrs) > 0 {
		return false, fmt.Errorf("failed to unmarshal %s: %w", clusterConfigurationFilename, errors.Join(strErrs...))
	}
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal %s: invalid JSON syntax", clusterConfigurationFilename)
	}

	cc.File, cc.Config = file, config
	if err = cc.finish(); err != nil {
		return false, err
	}

	return true, nil
}

func (cc *ClusterConfiguration) finish() error {
	if cc.Config == nil {
		return errors.New("missing configuration or manifest file")
	}
	return nil
}

func chronyConfWithAdditionalNTPSources(sources []string) string {
	content := defaultChronyConf[:]
	for _, source := range sources {
		content += fmt.Sprintf("\nserver %s iburst", source)
	}
	return content
}
