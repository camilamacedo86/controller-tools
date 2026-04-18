/*
Copyright 2026 The Kubernetes Authors.

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

package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"

	"sigs.k8s.io/controller-tools/pkg/applyconfiguration"
	"sigs.k8s.io/controller-tools/pkg/deepcopy"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("ApplyConfiguration auto-detection", func() {
	const testPackagePath = "../../pkg/applyconfiguration/testdata/cronjob/api/v1"

	setupRuntime := func(hasObjectGen, hasApplyConfigGen bool) (*genall.Runtime, int) {
		By("Initializing the marker registry")
		reg := &markers.Registry{}
		Expect(genall.RegisterOptionsMarkers(reg)).To(Succeed())

		deepcopyGen := deepcopy.Generator{}
		Expect(deepcopyGen.RegisterMarkers(reg)).To(Succeed())

		acGen := applyconfiguration.Generator{}
		Expect(acGen.RegisterMarkers(reg)).To(Succeed())

		By("Loading the test package")
		roots, err := loader.LoadRootsWithConfig(&packages.Config{}, testPackagePath)
		Expect(err).NotTo(HaveOccurred())

		col := &markers.Collector{Registry: reg}

		rt := &genall.Runtime{
			Generators: []*genall.Generator{},
			GenerationContext: genall.GenerationContext{
				Collector: col,
				Roots:     roots,
			},
		}

		if hasObjectGen {
			var gen genall.Generator = deepcopy.Generator{}
			rt.Generators = append(rt.Generators, &gen)
		}

		if hasApplyConfigGen {
			var gen genall.Generator = applyconfiguration.Generator{}
			rt.Generators = append(rt.Generators, &gen)
		}

		return rt, len(rt.Generators)
	}

	DescribeTable("should handle auto-detection correctly",
		func(rawOpts []string, hasObjectGen, hasApplyConfigGen, expectAdded bool) {
			rt, initialCount := setupRuntime(hasObjectGen, hasApplyConfigGen)

			By("Running auto-detection")
			err := enableApplyConfigIfMarked(rt, rawOpts)
			Expect(err).NotTo(HaveOccurred())

			finalCount := len(rt.Generators)
			wasAdded := finalCount > initialCount

			By("Verifying the generator was added or not as expected")
			Expect(wasAdded).To(Equal(expectAdded), "applyconfiguration generator addition mismatch")

			if expectAdded && wasAdded {
				lastGen := rt.Generators[len(rt.Generators)-1]
				_, ok := (*lastGen).(applyconfiguration.Generator)
				Expect(ok).To(BeTrue(), "last generator should be applyconfiguration.Generator")
			}
		},
		Entry("when applyconfiguration explicitly specified should not auto-enable",
			[]string{"applyconfiguration"}, true, false, false),
		Entry("when applyconfiguration already present should not add again",
			[]string{}, true, true, false),
		Entry("when no object generator should not auto-enable",
			[]string{}, false, false, false),
		Entry("when object generator and applyconfiguration markers indicate intent should auto-enable",
			[]string{}, true, false, true),
	)
})
