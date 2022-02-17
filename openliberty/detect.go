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

package openliberty

import (
	"fmt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/open-liberty/internal/server"
	"github.com/paketo-buildpacks/open-liberty/internal/util"
	"path/filepath"
)

const (
	PlanEntryOpenLiberty           = "open-liberty"
	PlanEntryJRE                   = "jre"
	PlanEntryJVMApplicationPackage = "jvm-application-package"
)

type Detect struct {
	Logger bard.Logger
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	cr, err := libpak.NewConfigurationResolver(context.Buildpack, &d.Logger)
	if err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("could not create configuration resolver\n%w", err)
	}
	serverName, _ := cr.Resolve("BP_OPENLIBERTY_SERVER_NAME")
	isPackagedServer, err :=
		util.FileExists(filepath.Join(context.Application.Path, "wlp", "usr", "servers", serverName, "server.xml"))
	if err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("unable to read packaged server.xml\n%w", err)
	}

	if isPackagedServer {
		return d.detectPackagedServer(context, serverName)
	}

	return d.detectApplication(context)
}

// detectApplication will handle detection of applications. It will pass detection iff `Main-Class` is not defined in
// the manifest. If a compiled artifact was pushed, detectApplication will mark the `jvm-application-package`
// requirement as being met.
func (d Detect) detectApplication(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	if mainClassDefined, err := util.ManifestHasMainClassDefined(context.Application.Path); err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("unable to check manifest\n%w", err)
	} else if mainClassDefined {
		return libcnb.DetectResult{Pass: false}, nil
	}

	// When a compiled artifact is pushed, mark that a JVM application package has been provided so that the build
	// plan requirement is satisfied.
	isJvmAppPackage, err := util.IsJvmApplicationPackage(context.Application.Path)
	if err != nil {
		return libcnb.DetectResult{}, err
	}

	result := libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: PlanEntryOpenLiberty},
				},

				Requires: []libcnb.BuildPlanRequire{
					{Name: PlanEntryJRE, Metadata: map[string]interface{}{
						"launch": true,
						"build":  true,
						"cache":  true},
					},
					{Name: PlanEntryJVMApplicationPackage},
					{Name: PlanEntryOpenLiberty},
				},
			},
		},
	}

	if isJvmAppPackage {
		result.Plans[0].Provides = append(result.Plans[0].Provides, libcnb.BuildPlanProvide{
			Name: PlanEntryJVMApplicationPackage,
		})
	}

	return result, nil
}

// detectPackagedServer handles detection of a packaged Liberty server.
func (d Detect) detectPackagedServer(context libcnb.DetectContext, serverName string) (libcnb.DetectResult, error) {
	libertyServer := server.LibertyServer{
		InstallRoot: filepath.Join(context.Application.Path, "wlp"),
		ServerName:  serverName,
	}
	hasApps, err := libertyServer.HasInstalledApps()
	if err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("unable to check if packaged server has apps\n%w", err)
	}

	result := libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: PlanEntryOpenLiberty},
				},

				Requires: []libcnb.BuildPlanRequire{
					{Name: PlanEntryJRE, Metadata: map[string]interface{}{
						"launch": true,
						"build":  true,
						"cache":  true},
					},
					{Name: PlanEntryJVMApplicationPackage},
					{Name: PlanEntryOpenLiberty, Metadata: map[string]interface{}{
						"packaged-server": true,
					}},
				},
			},
		},
	}

	if hasApps {
		result.Plans[0].Provides = append(result.Plans[0].Provides, libcnb.BuildPlanProvide{
			Name: PlanEntryJVMApplicationPackage,
		})
	}

	return result, nil
}
