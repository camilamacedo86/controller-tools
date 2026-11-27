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
	"errors"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The webhook and RBAC tests have no API server, so their generated
// manifests are checked here. Strict field validation makes the API server
// reject unknown fields instead of dropping them.
var _ = Describe("Generated webhook configurations", func() {
	It("should be accepted by the API server", func(ctx SpecContext) {
		paths, err := filepath.Glob("../../webhook/testdata/valid*/manifests.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).NotTo(BeEmpty())
		paths = append(paths, "../../webhook/testdata/manifests.yaml")

		for _, path := range paths {
			for _, obj := range readObjects(path) {
				err := k8sClient.Create(ctx, obj, client.DryRunAll, client.FieldValidation("Strict"))
				Expect(err).NotTo(HaveOccurred(), "%s: %s %s", path, obj.GetKind(), obj.GetName())
			}
		}
	})
})

var _ = Describe("Generated RBAC", func() {
	It("should be accepted by the API server", func(ctx SpecContext) {
		path := "../../rbac/testdata/role.yaml"
		for _, obj := range readObjects(path) {
			if obj.GetNamespace() != "" {
				namespace := &corev1.Namespace{}
				namespace.SetName(obj.GetNamespace())
				err := k8sClient.Create(ctx, namespace)
				Expect(client.IgnoreAlreadyExists(err)).NotTo(HaveOccurred())
			}
			err := k8sClient.Create(ctx, obj, client.DryRunAll, client.FieldValidation("Strict"))
			Expect(err).NotTo(HaveOccurred(), "%s: %s %s/%s", path, obj.GetKind(), obj.GetNamespace(), obj.GetName())
		}
	})
})

// readObjects reads every object in a multi-document YAML file.
func readObjects(path string) []*unstructured.Unstructured {
	GinkgoHelper()

	file, err := os.Open(path)
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	var objects []*unstructured.Unstructured
	decoder := yaml.NewYAMLOrJSONDecoder(file, 4096)
	for {
		obj := &unstructured.Unstructured{}
		err := decoder.Decode(&obj.Object)
		if errors.Is(err, io.EOF) {
			break
		}
		Expect(err).NotTo(HaveOccurred())
		if len(obj.Object) > 0 {
			objects = append(objects, obj)
		}
	}
	Expect(objects).NotTo(BeEmpty(), "no objects in %s", path)
	return objects
}
