package main

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/installer/cmd/openshift-install/command"
	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/cluster"
	"github.com/openshift/installer/pkg/asset/cluster/tfvars"
	"github.com/openshift/installer/pkg/asset/imagebased/configimage"
	"github.com/openshift/installer/pkg/asset/imagebased/image"
	"github.com/openshift/installer/pkg/asset/kubeconfig"
	"github.com/openshift/installer/pkg/asset/password"
	"github.com/openshift/installer/pkg/asset/tls"
	timer "github.com/openshift/installer/pkg/metrics/timer"
)

func newImageBasedCmd(ctx context.Context) *cobra.Command {
	imagebasedCmd := &cobra.Command{
		Use:   "image-based",
		Short: "Commands for supporting cluster installation using the image-based installer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	imagebasedCmd.AddCommand(newImageBasedCreateCmd(ctx))
	return imagebasedCmd
}

var (
	imageBasedInstallationConfigTemplateTarget = target{
		name: "Image-based Installation ISO Configuration template",
		command: &cobra.Command{
			Use:   "image-config-template",
			Short: "Generates a template of the Image-based Installation ISO config manifest used by the Image-based installer",
			Args:  cobra.ExactArgs(0),
		},
		assets: []asset.WritableAsset{
			&image.ImageBasedInstallationConfig{},
		},
	}

	imageBasedInstallationImageTarget = target{
		name: "Image-based Installation ISO Image",
		command: &cobra.Command{
			Use:   "image",
			Short: "Generates a bootable ISO image containing all the information needed to deploy a cluster",
			Args:  cobra.ExactArgs(0),
		},
		assets: []asset.WritableAsset{
			&image.Image{},
		},
	}

	imageBasedConfigTemplateTarget = target{
		name: "Image-based Installer Config ISO Configuration Template",
		command: &cobra.Command{
			Use:   "config-template",
			Short: "Generates a template of the Image-based Config ISO config manifest used by the image-based installer",
			Args:  cobra.ExactArgs(0),
		},
		assets: []asset.WritableAsset{
			&configimage.ImageBasedConfig{},
		},
	}

	imageBasedConfigImageTarget = target{
		name: "Image-based Installer Config ISO Image",
		command: &cobra.Command{
			Use:   "config-image",
			Short: "Generates an ISO containing configuration files only",
			Args:  cobra.ExactArgs(0),
		},
		assets: []asset.WritableAsset{
			&configimage.ConfigImage{},
			&kubeconfig.ImageBasedAdminClient{},
			&password.KubeadminPassword{},
		},
	}

	imageBasedAWSSingleNodeOpenShiftTarget = target{
		name: "Image-based Installer AWS Single Node OpenShift",
		command: &cobra.Command{
			Use:   "aws-sno",
			Short: "Creates an AWS Single Node OpenShift cluster based on the provided seed AMI",
			Args:  cobra.ExactArgs(0),
			PostRun: func(cmd *cobra.Command, _ []string) {
				ctx := cmd.Context()

				cleanup := command.SetupFileHook(command.RootOpts.Dir)
				defer cleanup()

				config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(command.RootOpts.Dir, "auth", "kubeconfig"))
				if err != nil {
					logrus.Fatal(errors.Wrap(err, "loading kubeconfig"))
				}
				client, err := kubernetes.NewForConfig(config)
				if err != nil {
					logrus.Fatal(err)
				}

				discovery := client.Discovery()
				apiTimeout := 20 * time.Minute
				untilTime := time.Now().Add(apiTimeout)
				timezone, _ := untilTime.Zone()
				logrus.Infof("Waiting up to %v (until %v %s) for the Kubernetes API at %s...",
					apiTimeout, untilTime.Format(time.Kitchen), timezone, config.Host)

				apiContext, cancel := context.WithTimeout(ctx, apiTimeout)
				defer cancel()
				// Poll quickly so we notice changes, but only log when the response
				// changes (because that's interesting) or when we've seen 15 of the
				// same errors in a row (to show we're still alive).
				logDownsample := 15
				silenceRemaining := logDownsample
				previousErrorSuffix := ""
				timer.StartTimer("API")

				var lastErr error
				err = wait.PollUntilContextCancel(apiContext, 2*time.Second, true, func(_ context.Context) (done bool, err error) {
					version, err := discovery.ServerVersion()
					if err == nil {
						logrus.Infof("API %s up", version)
						timer.StopTimer("API")
						return true, nil
					}

					lastErr = err
					silenceRemaining--
					chunks := strings.Split(err.Error(), ":")
					errorSuffix := chunks[len(chunks)-1]
					if previousErrorSuffix != errorSuffix {
						logrus.Debugf("Still waiting for the Kubernetes API: %v", err)
						previousErrorSuffix = errorSuffix
						silenceRemaining = logDownsample
					} else if silenceRemaining == 0 {
						logrus.Debugf("Still waiting for the Kubernetes API: %v", err)
						silenceRemaining = logDownsample
					}

					return false, nil
				})
				if err != nil {
					if lastErr != nil {
						logrus.Fatal(lastErr)
					}
					logrus.Fatal(err)
				}

				if err := waitForStableOperators(ctx, config); err != nil {
					logrus.Fatal(err)
				}

				consoleURL, err := getConsole(ctx, config)
				if err != nil {
					logrus.Warnf("Cluster does not have a console available: %v", err)
				}

				logComplete(command.RootOpts.Dir, consoleURL)
				timer.StopTimer(timer.TotalTimeElapsed)
				timer.LogSummary()
			},
		},
		assets: []asset.WritableAsset{
			&configimage.ClusterConfiguration{},
			&cluster.Metadata{},
			&tfvars.TerraformVariables{},
			&kubeconfig.ImageBasedAdminClient{},
			&password.KubeadminPassword{},
			&tls.JournalCertKey{},
			&cluster.Cluster{},
		},
	}

	imageBasedTargets = []target{
		imageBasedInstallationConfigTemplateTarget,
		imageBasedInstallationImageTarget,
		imageBasedConfigTemplateTarget,
		imageBasedConfigImageTarget,
		imageBasedAWSSingleNodeOpenShiftTarget,
	}
)

func newImageBasedCreateCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Commands for generating image-based installer artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	for _, t := range imageBasedTargets {
		t.command.Args = cobra.ExactArgs(0)
		t.command.Run = runTargetCmd(ctx, t.assets...)
		cmd.AddCommand(t.command)
	}

	return cmd
}
