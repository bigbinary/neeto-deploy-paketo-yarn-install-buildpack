package modules_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/yarn-cnb/modules"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

//go:generate mockgen -source=modules.go -destination=mocks_test.go -package=modules_test

func TestUnitModules(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Modules", testModules, spec.Report(report.Terminal{}))
}

func testModules(t *testing.T, when spec.G, it spec.S) {
	when("modules.NewContributor", func() {
		var (
			mockCtrl       *gomock.Controller
			mockPkgManager *MockPackageManager
			factory        *test.BuildFactory
		)

		it.Before(func() {
			mockCtrl = gomock.NewController(t)
			mockPkgManager = NewMockPackageManager(mockCtrl)

			factory = test.NewBuildFactory(t)
		})

		it.After(func() {
			mockCtrl.Finish()
		})

		when("there is no yarn.lock", func() {
			it("fails", func() {
				factory.AddBuildPlan(t, modules.Dependency, buildplan.Dependency{})

				_, _, err := modules.NewContributor(factory.Build, mockPkgManager)
				Expect(err).To(HaveOccurred())
			})
		})

		when("there is a yarn.lock", func() {
			it.Before(func() {
				layers.WriteToFile(
					strings.NewReader("yarn lock"),
					filepath.Join(factory.Build.Application.Root, "yarn.lock"),
					0666,
				)
			})

			it("returns true if a build plan exists", func() {
				factory.AddBuildPlan(t, modules.Dependency, buildplan.Dependency{})

				_, willContribute, err := modules.NewContributor(factory.Build, mockPkgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(willContribute).To(BeTrue())
			})

			it("returns false if a build plan does not exist", func() {
				_, willContribute, err := modules.NewContributor(factory.Build, mockPkgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(willContribute).To(BeFalse())
			})

			it("uses yarn.lock for identity", func() {
				factory.AddBuildPlan(t, modules.Dependency, buildplan.Dependency{})

				contributor, _, _ := modules.NewContributor(factory.Build, mockPkgManager)
				name, version := contributor.Identity()
				Expect(name).To(Equal(modules.Dependency))
				Expect(version).To(Equal("6a896d7017d636a532a914536a1cb7212c5d95a6ec5826d01e2b292e3a5d0a2a"))
			})

			when("the app is vendored", func() {
				it.Before(func() {
					layers.WriteToFile(
						strings.NewReader("some module"),
						filepath.Join(factory.Build.Application.Root, "node_modules", "test_module"),
						0666,
					)

					mockPkgManager.EXPECT().Rebuild(factory.Build.Application.Root)
				})

				it("contributes modules to the cache layer when included in the build plan", func() {
					factory.AddBuildPlan(t, modules.Dependency, buildplan.Dependency{
						Metadata: buildplan.Metadata{"build": true},
					})

					contributor, _, err := modules.NewContributor(factory.Build, mockPkgManager)
					Expect(err).NotTo(HaveOccurred())

					Expect(contributor.Contribute()).To(Succeed())

					layer := factory.Build.Layers.Layer(modules.Dependency)
					test.BeLayerLike(t, layer, true, true, false)
					test.BeFileLike(t, filepath.Join(layer.Root, "test_module"), 0644, "some module")
					test.BeOverrideSharedEnvLike(t, layer, "NODE_PATH", layer.Root)

					Expect(filepath.Join(factory.Build.Application.Root, "node_modules")).NotTo(BeADirectory())
				})

				it("contributes modules to the launch layer when included in the build plan", func() {
					factory.AddBuildPlan(t, modules.Dependency, buildplan.Dependency{
						Metadata: buildplan.Metadata{"launch": true},
					})

					contributor, _, err := modules.NewContributor(factory.Build, mockPkgManager)
					Expect(err).NotTo(HaveOccurred())

					Expect(contributor.Contribute()).To(Succeed())

					layer := factory.Build.Layers.Layer(modules.Dependency)
					test.BeLayerLike(t, layer, false, true, true)
					test.BeFileLike(t, filepath.Join(layer.Root, "test_module"), 0644, "some module")
					test.BeOverrideSharedEnvLike(t, layer, "NODE_PATH", layer.Root)

					Expect(filepath.Join(factory.Build.Application.Root, "node_modules")).NotTo(BeADirectory())
				})
			})

			when("the app is not vendored", func() {
				it.Before(func() {
					mockPkgManager.EXPECT().Install(factory.Build.Application.Root).Do(func(location string) {
						layers.WriteToFile(
							strings.NewReader("some module"),
							filepath.Join(factory.Build.Application.Root, "node_modules", "test_module"),
							0666,
						)
					})
				})

				it("contributes modules to the cache layer when included in the build plan", func() {
					factory.AddBuildPlan(t, modules.Dependency, buildplan.Dependency{
						Metadata: buildplan.Metadata{"build": true},
					})

					contributor, _, err := modules.NewContributor(factory.Build, mockPkgManager)
					Expect(err).NotTo(HaveOccurred())

					Expect(contributor.Contribute()).To(Succeed())

					layer := factory.Build.Layers.Layer(modules.Dependency)
					test.BeLayerLike(t, layer, true, true, false)
					test.BeFileLike(t, filepath.Join(layer.Root, "test_module"), 0644, "some module")
					test.BeOverrideSharedEnvLike(t, layer, "NODE_PATH", layer.Root)

					Expect(filepath.Join(factory.Build.Application.Root, "node_modules")).NotTo(BeADirectory())
				})

				it("contributes modules to the launch layer when included in the build plan", func() {
					factory.AddBuildPlan(t, modules.Dependency, buildplan.Dependency{
						Metadata: buildplan.Metadata{"launch": true},
					})

					contributor, _, err := modules.NewContributor(factory.Build, mockPkgManager)
					Expect(err).NotTo(HaveOccurred())

					Expect(contributor.Contribute()).To(Succeed())

					layer := factory.Build.Layers.Layer(modules.Dependency)
					test.BeLayerLike(t, layer, false, true, true)
					test.BeFileLike(t, filepath.Join(layer.Root, "test_module"), 0644, "some module")
					test.BeOverrideSharedEnvLike(t, layer, "NODE_PATH", layer.Root)

					Expect(filepath.Join(factory.Build.Application.Root, "node_modules")).NotTo(BeADirectory())
				})
			})
		})
	})
}
