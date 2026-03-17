// Package generator creates native platform projects from templates.
//
// For iOS, it generates an Xcode project with the native bridge code.
// Users never edit these files — they are regenerated on every build.
package generator

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/bridge_core.m templates/Info.plist templates/main.m templates/components/*.m
var templates embed.FS

// IOSConfig holds configuration for generating an iOS project.
type IOSConfig struct {
	AppName          string
	BundleID         string
	DeploymentTarget string
	OutputDir        string // the ios/ directory path
}

// GenerateIOS creates the iOS native project files.
func GenerateIOS(cfg IOSConfig) error {
	if cfg.AppName == "" {
		cfg.AppName = "GoxApp"
	}
	if cfg.BundleID == "" {
		cfg.BundleID = "com.gox." + strings.ToLower(cfg.AppName)
	}
	if cfg.DeploymentTarget == "" {
		cfg.DeploymentTarget = "16.0"
	}

	appDir := filepath.Join(cfg.OutputDir, cfg.AppName)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("creating app dir: %w", err)
	}

	// Clean up stale .m files from previous builds (e.g. old bridge.m)
	existingFiles, _ := os.ReadDir(appDir)
	for _, f := range existingFiles {
		if strings.HasSuffix(f.Name(), ".m") {
			os.Remove(filepath.Join(appDir, f.Name()))
		}
	}

	replacer := strings.NewReplacer(
		"{{APP_NAME}}", cfg.AppName,
		"{{BUNDLE_ID}}", cfg.BundleID,
		"{{EXECUTABLE_NAME}}", cfg.AppName,
		"{{DEPLOYMENT_TARGET}}", cfg.DeploymentTarget,
	)

	// Copy core template files
	coreFiles := []struct {
		template string
		output   string
	}{
		{"templates/bridge_core.m", filepath.Join(appDir, "bridge_core.m")},
		{"templates/Info.plist", filepath.Join(appDir, "Info.plist")},
		{"templates/main.m", filepath.Join(appDir, "main.m")},
	}

	for _, f := range coreFiles {
		data, err := templates.ReadFile(f.template)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", f.template, err)
		}
		content := replacer.Replace(string(data))
		if err := os.WriteFile(f.output, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", f.output, err)
		}
	}

	// Copy component .m files
	componentFiles, err := templates.ReadDir("templates/components")
	if err != nil {
		return fmt.Errorf("reading components dir: %w", err)
	}

	var componentNames []string
	for _, entry := range componentFiles {
		if strings.HasSuffix(entry.Name(), ".m") {
			componentNames = append(componentNames, entry.Name())
			data, err := templates.ReadFile("templates/components/" + entry.Name())
			if err != nil {
				return fmt.Errorf("reading component %s: %w", entry.Name(), err)
			}
			if err := os.WriteFile(filepath.Join(appDir, entry.Name()), data, 0644); err != nil {
				return fmt.Errorf("writing component %s: %w", entry.Name(), err)
			}
		}
	}

	// Generate the Xcode project
	if err := generateXcodeProject(cfg, appDir, componentNames); err != nil {
		return fmt.Errorf("generating Xcode project: %w", err)
	}

	return nil
}

func generateXcodeProject(cfg IOSConfig, appDir string, componentNames []string) error {
	projDir := filepath.Join(cfg.OutputDir, cfg.AppName+".xcodeproj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		return err
	}

	pbxproj := generatePbxproj(cfg, componentNames)
	return os.WriteFile(filepath.Join(projDir, "project.pbxproj"), []byte(pbxproj), 0644)
}

func generatePbxproj(cfg IOSConfig, componentNames []string) string {
	// All source .m files: bridge_core.m, main.m, plus component files
	allSources := []string{"bridge_core.m", "main.m"}
	allSources = append(allSources, componentNames...)

	// Generate dynamic PBX entries for source files
	// IDs: A1CCnnnn for build files, A2CCnnnn for file refs
	var buildFileSection strings.Builder
	var fileRefSection strings.Builder
	var sourcesBuildFiles strings.Builder
	var groupChildren strings.Builder

	for i, name := range allSources {
		buildID := fmt.Sprintf("A1CC%04d", i)
		refID := fmt.Sprintf("A2CC%04d", i)
		buildFileSection.WriteString(fmt.Sprintf(
			"\t\t%s /* %s in Sources */ = {isa = PBXBuildFile; fileRef = %s; };\n",
			buildID, name, refID))
		fileRefSection.WriteString(fmt.Sprintf(
			"\t\t%s /* %s */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.objc; path = %s; sourceTree = \"<group>\"; };\n",
			refID, name, name))
		sourcesBuildFiles.WriteString(fmt.Sprintf("\t\t\t\t%s /* %s in Sources */,\n", buildID, name))
		groupChildren.WriteString(fmt.Sprintf("\t\t\t\t%s /* %s */,\n", refID, name))
	}

	return fmt.Sprintf(`// !$*UTF8*$!
{
	archiveVersion = 1;
	classes = {
	};
	objectVersion = 56;
	objects = {

/* Begin PBXBuildFile section */
%[4]s		A1000003 /* libgox.a in Frameworks */ = {isa = PBXBuildFile; fileRef = A2000003; };
		A1000004 /* UIKit.framework in Frameworks */ = {isa = PBXBuildFile; fileRef = A2000004; };
		A1000005 /* Foundation.framework in Frameworks */ = {isa = PBXBuildFile; fileRef = A2000005; };
		A1000006 /* CoreGraphics.framework in Frameworks */ = {isa = PBXBuildFile; fileRef = A2000006; };
/* End PBXBuildFile section */

/* Begin PBXFileReference section */
%[5]s		A2000003 /* libgox.a */ = {isa = PBXFileReference; lastKnownFileType = archive.ar; path = libgox.a; sourceTree = "<group>"; };
		A2000004 /* UIKit.framework */ = {isa = PBXFileReference; lastKnownFileType = wrapper.framework; name = UIKit.framework; path = System/Library/Frameworks/UIKit.framework; sourceTree = SDKROOT; };
		A2000005 /* Foundation.framework */ = {isa = PBXFileReference; lastKnownFileType = wrapper.framework; name = Foundation.framework; path = System/Library/Frameworks/Foundation.framework; sourceTree = SDKROOT; };
		A2000006 /* CoreGraphics.framework */ = {isa = PBXFileReference; lastKnownFileType = wrapper.framework; name = CoreGraphics.framework; path = System/Library/Frameworks/CoreGraphics.framework; sourceTree = SDKROOT; };
		A2000007 /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
		A2000008 /* %[1]s.app */ = {isa = PBXFileReference; explicitFileType = wrapper.application; includeInIndex = 0; path = "%[1]s.app"; sourceTree = BUILT_PRODUCTS_DIR; };
/* End PBXFileReference section */

/* Begin PBXFrameworksBuildPhase section */
		A3000001 /* Frameworks */ = {
			isa = PBXFrameworksBuildPhase;
			buildActionMask = 2147483647;
			files = (
				A1000003 /* libgox.a in Frameworks */,
				A1000004 /* UIKit.framework in Frameworks */,
				A1000005 /* Foundation.framework in Frameworks */,
				A1000006 /* CoreGraphics.framework in Frameworks */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXFrameworksBuildPhase section */

/* Begin PBXGroup section */
		A4000001 = {
			isa = PBXGroup;
			children = (
				A4000002 /* %[1]s */,
				A4000003 /* Frameworks */,
				A4000004 /* Products */,
			);
			sourceTree = "<group>";
		};
		A4000002 /* %[1]s */ = {
			isa = PBXGroup;
			children = (
%[6]s				A2000007 /* Info.plist */,
			);
			path = "%[1]s";
			sourceTree = "<group>";
		};
		A4000003 /* Frameworks */ = {
			isa = PBXGroup;
			children = (
				A2000003 /* libgox.a */,
				A2000004 /* UIKit.framework */,
				A2000005 /* Foundation.framework */,
				A2000006 /* CoreGraphics.framework */,
			);
			name = Frameworks;
			sourceTree = "<group>";
		};
		A4000004 /* Products */ = {
			isa = PBXGroup;
			children = (
				A2000008 /* %[1]s.app */,
			);
			name = Products;
			sourceTree = "<group>";
		};
/* End PBXGroup section */

/* Begin PBXNativeTarget section */
		A5000001 /* %[1]s */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = A7000001;
			buildPhases = (
				A6000001 /* Sources */,
				A3000001 /* Frameworks */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = "%[1]s";
			productName = "%[1]s";
			productReference = A2000008;
			productType = "com.apple.product-type.application";
		};
/* End PBXNativeTarget section */

/* Begin PBXProject section */
		A8000001 /* Project object */ = {
			isa = PBXProject;
			buildConfigurationList = A7000002;
			compatibilityVersion = "Xcode 14.0";
			developmentRegion = en;
			hasScannedForEncodings = 0;
			knownRegions = (
				en,
				Base,
			);
			mainGroup = A4000001;
			productRefGroup = A4000004;
			projectDirPath = "";
			projectRoot = "";
			targets = (
				A5000001 /* %[1]s */,
			);
		};
/* End PBXProject section */

/* Begin PBXSourcesBuildPhase section */
		A6000001 /* Sources */ = {
			isa = PBXSourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
%[7]s			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXSourcesBuildPhase section */

/* Begin XCBuildConfiguration section */
		A9000001 /* Debug */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
				INFOPLIST_FILE = "%[1]s/Info.plist";
				IPHONEOS_DEPLOYMENT_TARGET = %[2]s;
				LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
				LIBRARY_SEARCH_PATHS = "$(inherited) $(PROJECT_DIR)";
				OTHER_LDFLAGS = "-lresolv";
				PRODUCT_BUNDLE_IDENTIFIER = "%[3]s";
				PRODUCT_NAME = "$(TARGET_NAME)";
				TARGETED_DEVICE_FAMILY = "1,2";
			};
			name = Debug;
		};
		A9000002 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
				INFOPLIST_FILE = "%[1]s/Info.plist";
				IPHONEOS_DEPLOYMENT_TARGET = %[2]s;
				LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
				LIBRARY_SEARCH_PATHS = "$(inherited) $(PROJECT_DIR)";
				OTHER_LDFLAGS = "-lresolv";
				PRODUCT_BUNDLE_IDENTIFIER = "%[3]s";
				PRODUCT_NAME = "$(TARGET_NAME)";
				TARGETED_DEVICE_FAMILY = "1,2";
			};
			name = Release;
		};
		A9000003 /* Debug */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ALWAYS_SEARCH_USER_PATHS = NO;
				CLANG_ENABLE_MODULES = YES;
				CLANG_ENABLE_OBJC_ARC = YES;
				CODE_SIGN_STYLE = Automatic;
				COPY_PHASE_STRIP = NO;
				DEBUG_INFORMATION_FORMAT = dwarf;
				GCC_DYNAMIC_NO_PIC = NO;
				GCC_OPTIMIZATION_LEVEL = 0;
				GCC_PREPROCESSOR_DEFINITIONS = "DEBUG=1";
				SDKROOT = iphoneos;
			};
			name = Debug;
		};
		A9000004 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ALWAYS_SEARCH_USER_PATHS = NO;
				CLANG_ENABLE_MODULES = YES;
				CLANG_ENABLE_OBJC_ARC = YES;
				CODE_SIGN_STYLE = Automatic;
				COPY_PHASE_STRIP = YES;
				DEBUG_INFORMATION_FORMAT = "dwarf-with-dsym";
				ENABLE_NS_ASSERTIONS = NO;
				GCC_OPTIMIZATION_LEVEL = s;
				SDKROOT = iphoneos;
				VALIDATE_PRODUCT = YES;
			};
			name = Release;
		};
/* End XCBuildConfiguration section */

/* Begin XCConfigurationList section */
		A7000001 /* Build configuration list for PBXNativeTarget */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				A9000001 /* Debug */,
				A9000002 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
		A7000002 /* Build configuration list for PBXProject */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				A9000003 /* Debug */,
				A9000004 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
/* End XCConfigurationList section */

	};
	rootObject = A8000001 /* Project object */;
}
`, cfg.AppName, cfg.DeploymentTarget, cfg.BundleID,
		buildFileSection.String(),  // [4]
		fileRefSection.String(),    // [5]
		groupChildren.String(),     // [6]
		sourcesBuildFiles.String()) // [7]
}
