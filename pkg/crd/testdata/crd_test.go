/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cronjob

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ = Describe("CronJob CRD", func() {
	It("should be successfully applied", func(ctx SpecContext) {
		data, err := os.ReadFile("testdata.kubebuilder.io_cronjobs.yaml")
		Expect(err).NotTo(HaveOccurred())

		crd := &apiextensionsv1.CustomResourceDefinition{}
		err = yaml.UnmarshalStrict(data, crd)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(ctx, crd)
		Expect(err).NotTo(HaveOccurred())
	})
})

// CRDs that the rest of this suite does not install. They use a dry run
// because some of them share a name. The v1beta1 files in gen/ are skipped
// because that API is no longer served.
var _ = Describe("Generated CRDs", func() {
	DescribeTable("should be accepted by the API server",
		func(ctx SpecContext, path string) {
			data, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())

			crd := &apiextensionsv1.CustomResourceDefinition{}
			err = yaml.UnmarshalStrict(data, crd)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(ctx, crd, client.DryRunAll)
			Expect(err).NotTo(HaveOccurred())
		},
		Entry("ExternalDocs", "testdata.kubebuilder.io_externaldocs.yaml"),
		Entry("Job", "testdata.kubebuilder.io_jobs.yaml"),
		Entry("Oneof", "testdata.kubebuilder.io_oneofs.yaml"),
		Entry("VersionedResource", "multiple_versions/testdata.kubebuilder.io_versionedresources.yaml"),
		Entry("TestQuota", "plural/plural.example.com_testquotas.yaml"),
		Entry("Foo", "gen/bar.example.com_foos.yaml"),
		Entry("Foo with maxDescLen", "gen/bar.example.com_foos_maxdesclen.yaml"),
		Entry("Zoo", "gen/zoo/bar.example.com_zoos.yaml"),
	)

	It("should reject the CronJob CRD generated without a group", func(ctx SpecContext) {
		data, err := os.ReadFile("testdata._cronjobs.yaml")
		Expect(err).NotTo(HaveOccurred())

		crd := &apiextensionsv1.CustomResourceDefinition{}
		err = yaml.UnmarshalStrict(data, crd)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(ctx, crd, client.DryRunAll)
		Expect(err).To(MatchError(ContainSubstring("spec.group: Required value")))
	})
})

var _ = Describe("Enum CRD", func() {
	It("should accept Value1", func(ctx SpecContext) {
		obj := enumObject("valid-value1", "Value1")
		err := k8sClient.Create(ctx, obj)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should accept Value2", func(ctx SpecContext) {
		obj := enumObject("valid-value2", "Value2")
		err := k8sClient.Create(ctx, obj)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should reject an invalid enum value", func(ctx SpecContext) {
		obj := enumObject("invalid-value", "Invalid")
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
	})
})

func enumObject(name, fieldValue string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "testdata.kubebuilder.io/v1",
			"kind":       "Enum",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": metav1.NamespaceDefault,
			},
			"spec": map[string]interface{}{
				"field": fieldValue,
			},
		},
	}
}
