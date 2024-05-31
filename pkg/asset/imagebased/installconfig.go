package imagebased

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/validation/field"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/installconfig"
	"github.com/openshift/installer/pkg/types"
	"github.com/openshift/installer/pkg/types/none"
	"github.com/openshift/installer/pkg/types/validation"
)

const (
	// InstallConfigFilename is the file containing the install-config.
	InstallConfigFilename = "install-config.yaml"
)

// OptionalInstallConfig is an InstallConfig where the default is empty, rather
// than generated from running the survey.
type OptionalInstallConfig struct {
	installconfig.AssetBase
	Supplied bool
}

var _ asset.WritableAsset = (*OptionalInstallConfig)(nil)

// Dependencies returns all of the dependencies directly needed by an
// InstallConfig asset.
func (a *OptionalInstallConfig) Dependencies() []asset.Asset {
	// Return no dependencies for the Agent install config, because it is
	// optional. We don't need to run the survey if it doesn't exist, since the
	// user may have supplied cluster-manifests that fully define the cluster.
	return []asset.Asset{}
}

// Generate generates the install-config.yaml file.
func (a *OptionalInstallConfig) Generate(parents asset.Parents) error {
	// Just generate an empty install config, since we have no dependencies.
	return nil
}

// Load returns the installconfig from disk.
func (a *OptionalInstallConfig) Load(f asset.FileFetcher) (bool, error) {
	found, err := a.LoadFromFile(f)
	if found && err == nil {
		a.Supplied = true
		if err := a.validateInstallConfig(a.Config).ToAggregate(); err != nil {
			return false, errors.Wrapf(err, "invalid install-config configuration")
		}
		if err := a.RecordFile(); err != nil {
			return false, err
		}
	}
	return found, err
}

func (a *OptionalInstallConfig) validateInstallConfig(installConfig *types.InstallConfig) field.ErrorList {
	var allErrs field.ErrorList
	if err := validation.ValidateInstallConfig(a.Config, true); err != nil {
		allErrs = append(allErrs, err...)
	}

	if err := a.validateSupportedPlatforms(installConfig); err != nil {
		allErrs = append(allErrs, err...)
	}

	if installConfig.FeatureSet != configv1.Default {
		allErrs = append(allErrs, field.NotSupported(field.NewPath("FeatureSet"), installConfig.FeatureSet, []string{string(configv1.Default)}))
	}

	warnUnusedConfig(installConfig)

	if err := a.validateSNOConfiguration(installConfig); err != nil {
		allErrs = append(allErrs, err...)
	}

	return allErrs
}

func (a *OptionalInstallConfig) validateSupportedPlatforms(installConfig *types.InstallConfig) field.ErrorList {
	var allErrs field.ErrorList

	fieldPath := field.NewPath("Platform")

	if installConfig.Platform.Name() != "" && installConfig.Platform.Name() != none.Name {
		allErrs = append(allErrs, field.NotSupported(fieldPath, installConfig.Platform.Name(), []string{none.Name}))
	}

	return allErrs
}

func (a *OptionalInstallConfig) validateSNOConfiguration(installConfig *types.InstallConfig) field.ErrorList {
	var allErrs field.ErrorList
	var fieldPath *field.Path

	controlPlaneReplicas := *installConfig.ControlPlane.Replicas
	if installConfig.ControlPlane != nil && controlPlaneReplicas != 1 {
		fieldPath = field.NewPath("ControlPlane", "Replicas")
		allErrs = append(allErrs, field.Required(fieldPath, fmt.Sprintf("Only Single Node OpenShift (SNO) is supported, total number of ControlPlane.Replicas must be 1. Found %v", controlPlaneReplicas)))
	}

	var workers int
	for _, worker := range installConfig.Compute {
		workers += int(*worker.Replicas)
	}
	if workers != 0 {
		fieldPath = field.NewPath("Compute", "Replicas")
		allErrs = append(allErrs, field.Required(fieldPath, fmt.Sprintf("Total number of Compute.Replicas must be 0 when ControlPlane.Replicas is 1 for platform %s. Found %v", none.Name, workers)))
	}

	if installConfig.Networking.NetworkType != "OVNKubernetes" {
		fieldPath = field.NewPath("Networking", "NetworkType")
		allErrs = append(allErrs, field.Invalid(fieldPath, installConfig.Networking.NetworkType, "Only OVNKubernetes network type is allowed for Single Node OpenShift (SNO) cluster"))
	}

	machineNetworksCount := len(installConfig.Networking.MachineNetwork)
	if machineNetworksCount != 1 {
		fieldPath = field.NewPath("Networking", "MachineNetwork")
		allErrs = append(allErrs, field.TooMany(fieldPath, machineNetworksCount, 1))
	}

	return allErrs
}

// ClusterName returns the name of the cluster, or a default name if no
// InstallConfig is supplied.
func (a *OptionalInstallConfig) ClusterName() string {
	if a.Config != nil && a.Config.ObjectMeta.Name != "" {
		return a.Config.ObjectMeta.Name
	}
	return "imagebased-sno-cluster"
}

// ClusterNamespace returns the namespace of the cluster.
func (a *OptionalInstallConfig) ClusterNamespace() string {
	if a.Config != nil && a.Config.ObjectMeta.Namespace != "" {
		return a.Config.ObjectMeta.Namespace
	}
	return ""
}

func warnUnusedConfig(installConfig *types.InstallConfig) {
	// "Proxyonly" is the default set from generic install config code
	if installConfig.AdditionalTrustBundlePolicy != "Proxyonly" {
		fieldPath := field.NewPath("AdditionalTrustBundlePolicy")
		logrus.Warnf(fmt.Sprintf("%s: %s is ignored", fieldPath, installConfig.AdditionalTrustBundlePolicy))
	}

	for i, compute := range installConfig.Compute {
		if compute.Hyperthreading != "Enabled" {
			fieldPath := field.NewPath(fmt.Sprintf("Compute[%d]", i), "Hyperthreading")
			logrus.Warnf(fmt.Sprintf("%s: %s is ignored", fieldPath, compute.Hyperthreading))
		}

		if compute.Platform != (types.MachinePoolPlatform{}) {
			fieldPath := field.NewPath(fmt.Sprintf("Compute[%d]", i), "Platform")
			logrus.Warnf(fmt.Sprintf("%s is ignored", fieldPath))
		}
	}

	if installConfig.ControlPlane.Hyperthreading != "Enabled" {
		fieldPath := field.NewPath("ControlPlane", "Hyperthreading")
		logrus.Warnf(fmt.Sprintf("%s: %s is ignored", fieldPath, installConfig.ControlPlane.Hyperthreading))
	}

	if installConfig.ControlPlane.Platform != (types.MachinePoolPlatform{}) {
		fieldPath := field.NewPath("ControlPlane", "Platform")
		logrus.Warnf(fmt.Sprintf("%s is ignored", fieldPath))
	}
}
