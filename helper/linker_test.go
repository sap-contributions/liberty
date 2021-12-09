/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package helper_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/open-liberty/helper"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLink(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		linker helper.FileLinker

		appDir   string
		layerDir string
	)

	it.Before(func() {
		var err error

		appDir, err = ioutil.TempDir("", "execd-helper-apps")
		Expect(err).NotTo(HaveOccurred())

		layerDir, err = ioutil.TempDir("", "execd-helper-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(appDir)).To(Succeed())
		Expect(os.RemoveAll(layerDir)).To(Succeed())
	})

	it("fails with default values because local directories do not exist", func() {
		Expect("/workspace").NotTo(BeADirectory())
		Expect("/layers/paketo-buildpacks_open-liberty/open-liberty-runtime").NotTo(BeADirectory())

		_, err := linker.Execute()
		Expect(err).To(HaveOccurred())
	})

	context("with explicit env vars set to valid dirs", func() {
		it.Before(func() {
			Expect(os.Setenv("BPI_OL_DROPIN_DIR", appDir)).To(Succeed())
			Expect(os.Setenv("BPI_OL_RUNTIME_ROOT", layerDir)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(layerDir, "usr", "servers", "defaultServer", "dropins"), 0755)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BPI_OL_DROPIN_DIR")).To(Succeed())
			Expect(os.Unsetenv("BPI_OL_RUNTIME_ROOT")).To(Succeed())

			Expect(os.RemoveAll(filepath.Join(layerDir, "usr", "servers", "defaultServer", "dropins"))).To(Succeed())
		})

		it("works", func() {
			_, err := linker.Execute()
			Expect(err).NotTo(HaveOccurred())

			resolvedAppDir, err := filepath.EvalSymlinks(appDir)
			Expect(err).NotTo(HaveOccurred())

			linkName := filepath.Join(layerDir, "usr", "servers", "defaultServer", "dropins", filepath.Base(appDir))
			resolved, err := filepath.EvalSymlinks(linkName)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved).To(Equal(resolvedAppDir))
		})
	})
}
