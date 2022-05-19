package tektonrun_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/pkg/reconciler/tektonrun"
)

var _ = Describe("Validate Tekton Run", func() {

	var tektonRun *tektonv1alpha1.Run

	BeforeEach(func() {
		tektonRun = &tektonv1alpha1.Run{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "build-pipeline",
				Name:      "image-build",
			},
		}
	})

	It("is valid if the Run references a Shipwright Build", func() {
		tektonRun.Spec = tektonv1alpha1.RunSpec{
			Ref: &tektonv1beta1.TaskRef{
				Kind:       "Build",
				APIVersion: "shipwright.io/v1alpha1",
				Name:       "image-build",
			},
		}
		err := tektonrun.ValidateTektonRun(tektonRun)
		Expect(err).NotTo(HaveOccurred())
	})

	It("is valid if the Run has a valid embedded task spec", func() {
		sourceURL := "https://github.com/shipwright-io/build"
		revision := "main"
		buildKind := buildv1alpha1.BuildStrategyKind("BuildStrategy")
		buildRunSpec := buildv1alpha1.BuildSpec{
			Source: buildv1alpha1.Source{
				URL:      &sourceURL,
				Revision: &revision,
			},
			Strategy: buildv1alpha1.Strategy{
				Kind: &buildKind,
				Name: "kaniko",
			},
			Output: buildv1alpha1.Image{
				Image: "ghcr.io/shipwright-io/build/shipwright-build-controller:latest",
			},
		}
		rawBuildRun, err := json.Marshal(buildRunSpec)
		Expect(err).NotTo(HaveOccurred())
		tektonRun.Spec = tektonv1alpha1.RunSpec{
			Spec: &tektonv1alpha1.EmbeddedRunSpec{
				TypeMeta: runtime.TypeMeta{
					APIVersion: "shipwright.io/v1alpha1",
					Kind:       "Build",
				},
				Spec: runtime.RawExtension{
					Raw: rawBuildRun,
				},
			},
		}
		err = tektonrun.ValidateTektonRun(tektonRun)
		Expect(err).NotTo(HaveOccurred())
	})

	It("is invalid if the Run reference has an incorrect APIVersion and Kind", func() {
		tektonRun.Spec = tektonv1alpha1.RunSpec{
			Ref: &tektonv1beta1.TaskRef{
				Kind:       "Bad",
				APIVersion: "something.awful.io/v1",
			},
		}
		err := tektonrun.ValidateTektonRun(tektonRun)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("kind must be Build"))
		Expect(err.Error()).To(ContainSubstring("apiVersion must be shipwright.io/v1alpha1"))
	})

	It("is invalid if the Run embedded spec has an incorrect APIVersion and Kind", func() {
		tektonRun.Spec = tektonv1alpha1.RunSpec{
			Spec: &tektonv1alpha1.EmbeddedRunSpec{
				TypeMeta: runtime.TypeMeta{
					APIVersion: "something.awful.io",
					Kind:       "Bad",
				},
			},
		}
		err := tektonrun.ValidateTektonRun(tektonRun)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("kind must be Build"))
		Expect(err.Error()).To(ContainSubstring("apiVersion must be shipwright.io/v1alpha1"))
	})

	It("is invalid if the Run has an unnamed task reference", func() {
		tektonRun.Spec = tektonv1alpha1.RunSpec{
			Ref: &tektonv1beta1.TaskRef{
				Kind:       "Build",
				APIVersion: "shipwright.io/v1alpha1",
			},
		}
		err := tektonrun.ValidateTektonRun(tektonRun)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("build name is required"))
	})

	It("is valid if the Run has a timeout", func() {
		tektonRun.Spec = tektonv1alpha1.RunSpec{
			Timeout: &metav1.Duration{
				Duration: 1 * time.Hour,
			},
			Ref: &tektonv1beta1.TaskRef{
				Kind:       "Build",
				APIVersion: "shipwright.io/v1alpha1",
				Name:       "image-build",
			},
		}
		err := tektonrun.ValidateTektonRun(tektonRun)
		Expect(err).NotTo(HaveOccurred())
	})

	It("is invalid if the Run has retries specified", func() {
		tektonRun.Spec = tektonv1alpha1.RunSpec{
			Retries: 3,
		}
		err := tektonrun.ValidateTektonRun(tektonRun)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("retries are not supported"))

	})

	It("is invalid if the Run has arbitrary parameters specified", func() {
		tektonRun.Spec = tektonv1alpha1.RunSpec{
			Params: []tektonv1beta1.Param{
				{
					Name:  "abrbitrary-apram",
					Value: *tektonv1beta1.NewArrayOrString("arbitrary-value"),
				},
			},
		}
		err := tektonrun.ValidateTektonRun(tektonRun)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unsupported value:"))
	})

	It("is valid if the Run has correct parameters specified", func() {
		tektonRun.Spec = tektonv1alpha1.RunSpec{
			Params: []tektonv1beta1.Param{
				{
					Name:  "shp-source-url",
					Value: *tektonv1beta1.NewArrayOrString("https://github.com/shipwright-io/build"),
				},
				{
					Name:  "shp-source-revision",
					Value: *tektonv1beta1.NewArrayOrString("main"),
				},
				{
					Name:  "shp-output-image",
					Value: *tektonv1beta1.NewArrayOrString("ghcr.io/shipwright-io/build/shipwright-build-controller:latest"),
				},
			},
			Ref: &tektonv1beta1.TaskRef{
				Kind:       "Build",
				APIVersion: "shipwright.io/v1alpha1",
				Name:       "image-build",
			},
		}
		err := tektonrun.ValidateTektonRun(tektonRun)
		Expect(err).NotTo(HaveOccurred())
	})
})