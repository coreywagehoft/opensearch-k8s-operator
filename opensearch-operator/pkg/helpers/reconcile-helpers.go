package helpers

import (
	"fmt"
	"path"

	"github.com/hashicorp/go-version"
	"k8s.io/utils/pointer"
	opsterv1 "opensearch.opster.io/api/v1"
)

func ResolveInitHelperImage(cr *opsterv1.OpenSearchCluster) (result opsterv1.ImageSpec) {
	defaultRepo := "public.ecr.aws/opsterio"
	defaultImage := "busybox"
	defaultVersion := "1.27.2-buildx"

	// If a custom InitHelper image is specified, use it.
	if cr.Spec.InitHelper.ImageSpec != nil {
		if useCustomImage(cr.Spec.InitHelper.ImageSpec, &result) {
			return
		}
	}

	// If a different image repo is requested, use that with the default image name and version tag.
	if cr.Spec.General.DefaultRepo != nil {
		defaultRepo = *cr.Spec.General.DefaultRepo
	}

	if cr.Spec.InitHelper.Version != nil {
		defaultVersion = *cr.Spec.InitHelper.Version
	}

	result.Image = pointer.String(fmt.Sprintf("%s:%s",
		path.Join(defaultRepo, defaultImage), defaultVersion))
	return
}

func ResolveImage(cr *opsterv1.OpenSearchCluster, nodePool *opsterv1.NodePool) (result opsterv1.ImageSpec) {
	defaultRepo := "docker.io/opensearchproject"
	defaultImage := "opensearch"

	var version string

	// If a general custom image is specified, use it.
	if cr.Spec.General.ImageSpec != nil {
		if useCustomImage(cr.Spec.General.ImageSpec, &result) {
			return
		}
	}

	// Calculate version based on upgrading status
	if nodePool == nil {
		version = cr.Spec.General.Version
	} else {
		componentStatus := opsterv1.ComponentStatus{
			Component:   "Upgrader",
			Description: nodePool.Component,
		}
		_, found := FindFirstPartial(cr.Status.ComponentsStatus, componentStatus, GetByDescriptionAndGroup)

		if cr.Status.Version == "" || cr.Status.Version == cr.Spec.General.Version {
			version = cr.Spec.General.Version
		} else {
			if found {
				version = cr.Spec.General.Version
			} else {
				version = cr.Status.Version
			}
		}
	}

	// If a different image repo is requested, use that with the default image
	// name and version tag.
	if cr.Spec.General.DefaultRepo != nil {
		defaultRepo = *cr.Spec.General.DefaultRepo
	}

	result.Image = pointer.String(fmt.Sprintf("%s:%s",
		path.Join(defaultRepo, defaultImage), version))
	return
}

func ResolveDashboardsImage(cr *opsterv1.OpenSearchCluster) (result opsterv1.ImageSpec) {
	defaultRepo := "docker.io/opensearchproject"
	defaultImage := "opensearch-dashboards"

	// If a custom dashboard image is specified, use it.
	if cr.Spec.Dashboards.ImageSpec != nil {
		if useCustomImage(cr.Spec.Dashboards.ImageSpec, &result) {
			return
		}
	}

	// If a different image repo is requested, use that with the default image
	// name and version tag.
	if cr.Spec.General.DefaultRepo != nil {
		defaultRepo = *cr.Spec.General.DefaultRepo
	}

	result.Image = pointer.String(fmt.Sprintf("%s:%s",
		path.Join(defaultRepo, defaultImage), cr.Spec.Dashboards.Version))
	return
}

func useCustomImage(customImageSpec *opsterv1.ImageSpec, result *opsterv1.ImageSpec) bool {
	if customImageSpec != nil {
		if customImageSpec.ImagePullPolicy != nil {
			result.ImagePullPolicy = customImageSpec.ImagePullPolicy
		}
		if len(customImageSpec.ImagePullSecrets) > 0 {
			result.ImagePullSecrets = customImageSpec.ImagePullSecrets
		}
		if customImageSpec.Image != nil {
			// If custom image is specified, use it.
			result.Image = customImageSpec.Image
			return true
		}
	}
	return false
}

// Function to help identify httpPort and securityconfigPath for 1.x and 2.x OpenSearch Operator.
func VersionCheck(instance *opsterv1.OpenSearchCluster) (int32, string) {
	var httpPort int32
	var securityconfigPath string
	versionPassed, _ := version.NewVersion(instance.Spec.General.Version)
	constraints, _ := version.NewConstraint(">= 2.0")
	if constraints.Check(versionPassed) {
		if instance.Spec.General.HttpPort > 0 {
			httpPort = instance.Spec.General.HttpPort
		} else {
			httpPort = 9200
		}
		securityconfigPath = "/usr/share/opensearch/config/opensearch-security"
	} else {
		httpPort = 9300
		securityconfigPath = "/usr/share/opensearch/plugins/opensearch-security/securityconfig"
	}
	return httpPort, securityconfigPath
}
