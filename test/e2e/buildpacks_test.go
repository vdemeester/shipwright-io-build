package e2e

import (
	operator "github.com/redhat-developer/build/pkg/apis/build/v1alpha1"
)

// buildpacks-v3 Test data setup
func buildpackBuildTestData(ns string, identifier string) (*operator.ClusterBuildStrategy, *operator.Build, *operator.BuildRun, error) {
	buildstrategy, err := clusterBuildStrategyTestData("samples/buildstrategy/buildpacks-v3/buildstrategy_buildpacks-v3_cr.yaml")
	if err != nil {
		return nil, nil, nil, err
	}

	build, err := buildTestData(ns, identifier, "samples/build/build_buildpacks-v3_cr.yaml")
	if err != nil {
		return nil, nil, nil, err
	}

	buildRun, err := buildRunTestData(ns, identifier, "samples/buildrun/buildrun_buildpacks-v3_cr.yaml")
	if err != nil {
		return nil, nil, nil, err
	}

	return buildstrategy, build, buildRun, nil
}
