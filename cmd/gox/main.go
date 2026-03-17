// Command gox is the GOX compiler and build tool CLI.
//
// Usage:
//
//	gox compile <file.gox|dir>    Compile .gox files to .go
//	gox generate ios              Generate iOS project in ios/
//	gox run ios                   Build and run on iOS simulator
//	gox run ios --device          Pick a simulator device interactively
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gox/internal/compiler"
	"gox/internal/generator"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "compile":
		err = cmdCompile()
	case "generate":
		err = cmdGenerate()
	case "run":
		err = cmdRun()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: gox <command> [args]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  compile <file.gox|dir>    Compile .gox files to .go")
	fmt.Fprintln(os.Stderr, "  generate ios              Generate iOS native project")
	fmt.Fprintln(os.Stderr, "  run ios [flags]           Build and run on iOS simulator")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "run flags:")
	fmt.Fprintln(os.Stderr, "  --device, -d              Pick simulator device interactively")
	fmt.Fprintln(os.Stderr, "  --logs, -l                Stream app logs after launch (Ctrl+C to stop)")
}

// --- compile ---

func cmdCompile() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: gox compile <file.gox|dir>")
	}
	return runCompile(os.Args[2])
}

func runCompile(target string) error {
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("cannot access %s: %w", target, err)
	}

	if info.IsDir() {
		return compileDir(target)
	}
	return compileFile(target)
}

func compileFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	result := compiler.Compile(src, path)

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintln(os.Stderr, e)
		}
		return fmt.Errorf("compilation failed with %d error(s)", len(result.Errors))
	}

	outPath := outputPath(path)
	if err := os.WriteFile(outPath, []byte(result.GoSource), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	fmt.Printf("  %s → %s\n", path, outPath)
	return nil
}

func compileDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	compiled := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".gox") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := compileFile(path); err != nil {
			return err
		}
		compiled++
	}

	if compiled == 0 {
		fmt.Println("no .gox files found")
	}
	return nil
}

func outputPath(goxPath string) string {
	ext := filepath.Ext(goxPath)
	base := strings.TrimSuffix(goxPath, ext)
	return base + "_gox.go"
}

// --- generate ---

func cmdGenerate() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: gox generate <ios|android>")
	}

	switch os.Args[2] {
	case "ios":
		return generateIOS()
	default:
		return fmt.Errorf("unknown platform: %s (supported: ios)", os.Args[2])
	}
}

func generateIOS() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	appName := filepath.Base(cwd)
	iosDir := filepath.Join(cwd, "ios")

	return generator.GenerateIOS(generator.IOSConfig{
		AppName:   appName,
		OutputDir: iosDir,
	})
}

// --- run ---

func cmdRun() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: gox run <ios|android>")
	}

	switch os.Args[2] {
	case "ios":
		return runIOS()
	default:
		return fmt.Errorf("unknown platform: %s (supported: ios)", os.Args[2])
	}
}

func runIOS() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	appName := filepath.Base(cwd)
	iosDir := filepath.Join(cwd, "ios")

	// Resolve device before building (fail fast if no simulators)
	interactive := hasFlag("--device", "-d")
	device, err := resolveDevice(interactive)
	if err != nil {
		return err
	}
	fmt.Printf("Target: %s (%s)\n\n", device.Name, device.Runtime)

	// Step 1: Compile all .gox files
	fmt.Println("[1/5] Compiling .gox files...")
	if err := compileAllGox(cwd); err != nil {
		return fmt.Errorf("compile: %w", err)
	}

	// Step 2: Generate bootstrap
	fmt.Println("[2/5] Generating bootstrap...")
	if err := generator.GenerateBootstrap(cwd); err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}

	// Step 3: Generate iOS project
	fmt.Println("[3/5] Generating iOS project...")
	if err := generator.GenerateIOS(generator.IOSConfig{
		AppName:   appName,
		OutputDir: iosDir,
	}); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	// Step 4: Build Go as C archive
	fmt.Println("[4/5] Building Go → C archive...")
	if err := buildGoArchive(cwd, iosDir); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	// Step 5: Build native app, install, launch
	fmt.Println("[5/5] Building and launching...")
	streamLogs := hasFlag("--logs", "-l")
	return buildAndLaunch(cwd, iosDir, appName, device, streamLogs)
}

func compileAllGox(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "ios" || base == "android" || base == "build" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".gox") {
			return compileFile(path)
		}
		return nil
	})
}

func buildGoArchive(projectDir, iosDir string) error {
	outPath := filepath.Join(iosDir, "libgox.a")
	sdkPath := iosSimulatorSDK()

	// Find the gox module root (where the yoga lib lives)
	goxRoot := findGoxRoot(projectDir)

	// Build libyoga.a for iOS simulator (separate from host build)
	yogaLibDir := filepath.Join(goxRoot, "internal", "yoga", "lib")
	if err := buildYogaForIOS("", yogaLibDir, sdkPath, iosDir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: yoga iOS build: %v\n", err)
	}

	cmd := exec.Command("go", "build",
		"-buildmode=c-archive",
		"-o", outPath,
		".",
	)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(),
		"GOOS=ios",
		"GOARCH=arm64",
		"CGO_ENABLED=1",
		"CC=clang",
		fmt.Sprintf("CGO_CFLAGS=-isysroot %s -miphonesimulator-version-min=16.0 -arch arm64", sdkPath),
		fmt.Sprintf("CGO_LDFLAGS=-isysroot %s -miphonesimulator-version-min=16.0 -arch arm64 -L%s -lyoga -lc++",
			sdkPath, yogaLibDir),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// buildYogaForIOS compiles Yoga C++ sources for the iOS simulator target.
// Outputs to iosDir (NOT the host lib dir) to avoid overwriting the macOS build.
func buildYogaForIOS(yogaSrcDir, yogaLibDir, sdkPath, iosDir string) error {
	yogaInclude := filepath.Join(yogaLibDir, "include")
	if _, err := os.Stat(yogaInclude); os.IsNotExist(err) {
		return fmt.Errorf("yoga include dir not found: %s", yogaInclude)
	}

	// Check if iOS yoga lib already exists
	iosYogaLib := filepath.Join(iosDir, "libyoga_ios.a")
	if _, err := os.Stat(iosYogaLib); err == nil {
		return nil // already built
	}

	var cppFiles []string
	filepath.Walk(yogaInclude, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(path, ".cpp") {
			cppFiles = append(cppFiles, path)
		}
		return nil
	})

	if len(cppFiles) == 0 {
		return fmt.Errorf("no yoga .cpp files found")
	}

	// Build dir for iOS .o files
	iosBuildDir := filepath.Join(iosDir, "yoga_build")
	os.MkdirAll(iosBuildDir, 0755)

	var objFiles []string
	for _, cpp := range cppFiles {
		// Output .o to iosBuildDir to avoid polluting source tree
		objName := strings.ReplaceAll(cpp, "/", "_")
		objName = strings.TrimSuffix(objName, ".cpp") + ".o"
		obj := filepath.Join(iosBuildDir, objName)

		cmd := exec.Command("clang++",
			"-std=c++20", "-O2", "-fPIC",
			"-isysroot", sdkPath,
			"-miphonesimulator-version-min=16.0",
			"-arch", "arm64",
			"-I"+yogaInclude,
			"-c", cpp,
			"-o", obj,
		)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("compiling %s: %w", filepath.Base(cpp), err)
		}
		objFiles = append(objFiles, obj)
	}

	args := append([]string{"rcs", iosYogaLib}, objFiles...)
	cmd := exec.Command("ar", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// findGoxRoot finds the gox module root by looking for the go.mod with "module gox".
func findGoxRoot(projectDir string) string {
	// Check if this IS the gox module
	modFile := filepath.Join(projectDir, "go.mod")
	if data, err := os.ReadFile(modFile); err == nil {
		if strings.Contains(string(data), "replace gox =>") {
			// User project with replace directive — extract the path
			for _, line := range strings.Split(string(data), "\n") {
				if strings.Contains(line, "replace gox =>") {
					parts := strings.Split(line, "=>")
					if len(parts) == 2 {
						path := strings.TrimSpace(parts[1])
						if !filepath.IsAbs(path) {
							path = filepath.Join(projectDir, path)
						}
						return path
					}
				}
			}
		}
		if strings.Contains(string(data), "module gox") {
			return projectDir
		}
	}
	return projectDir
}

func buildAndLaunch(projectDir, iosDir, appName string, device simDevice, streamLogs bool) error {
	sdkPath := iosSimulatorSDK()
	bundleID := "com.gox." + strings.ToLower(appName)

	// Build .app bundle
	appDir := filepath.Join(iosDir, "build", appName+".app")
	os.MkdirAll(appDir, 0755)

	// Copy Info.plist
	plistData, err := os.ReadFile(filepath.Join(iosDir, appName, "Info.plist"))
	if err != nil {
		return fmt.Errorf("reading Info.plist: %w", err)
	}
	os.WriteFile(filepath.Join(appDir, "Info.plist"), plistData, 0644)

	// Compile bridge.m + link with libgox.a + libyoga_ios.a
	clangArgs := []string{
		"-isysroot", sdkPath,
		"-miphonesimulator-version-min=16.0",
		"-arch", "arm64",
		"-framework", "UIKit",
		"-framework", "Foundation",
		"-framework", "CoreGraphics",
		"-framework", "Security",
		"-I", iosDir,
		filepath.Join(iosDir, appName, "bridge.m"),
		filepath.Join(iosDir, "libgox.a"),
		filepath.Join(iosDir, "libyoga_ios.a"),
		"-lc++",
		"-lresolv",
		"-o", filepath.Join(appDir, appName),
	}

	cmd := exec.Command("clang", clangArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clang: %w", err)
	}

	// Open Simulator.app
	_ = exec.Command("open", "-a", "Simulator").Run()

	// Boot the device and wait until ready
	boot := exec.Command("xcrun", "simctl", "boot", device.UDID)
	boot.Run() // ignore "already booted"

	// Wait for the device to finish booting
	waitBoot := exec.Command("xcrun", "simctl", "bootstatus", device.UDID, "-b")
	waitBoot.Stdout = os.Stdout
	waitBoot.Run()

	// Terminate previous instance if running
	_ = exec.Command("xcrun", "simctl", "terminate", device.UDID, bundleID).Run()

	// Install
	install := exec.Command("xcrun", "simctl", "install", device.UDID, appDir)
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	// Launch
	fmt.Printf("\n  Launching %s on %s...\n\n", appName, device.Name)
	launch := exec.Command("xcrun", "simctl", "launch", device.UDID, bundleID)
	launch.Stdout = os.Stdout
	launch.Stderr = os.Stderr
	if err := launch.Run(); err != nil {
		return err
	}

	if !streamLogs {
		return nil
	}

	// Stream logs — filter to our process, show GOX: prefixed lines
	fmt.Println("  Streaming logs (Ctrl+C to stop)...")
	logCmd := exec.Command("xcrun", "simctl", "spawn", device.UDID,
		"log", "stream",
		"--predicate", fmt.Sprintf("process == \"%s\"", appName),
		"--style", "compact",
	)

	stdout, err := logCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("log stream: %w", err)
	}
	logCmd.Stderr = os.Stderr

	if err := logCmd.Start(); err != nil {
		return fmt.Errorf("log stream start: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		// Only show user logs (slog output), not internal bridge debug logs
		if idx := strings.Index(line, "GOX_LOG: "); idx != -1 {
			fmt.Println(line[idx+9:])
		}
	}

	return logCmd.Wait()
}

// --- Simulator device management ---

type simDevice struct {
	Name      string `json:"name"`
	UDID      string `json:"udid"`
	State     string `json:"state"`
	Runtime   string // e.g. "iOS 26.2"
	Available bool   `json:"isAvailable"`
}

// resolveDevice picks a simulator device.
// If interactive, shows a list and lets the user choose.
// Otherwise, picks the first available iPhone automatically.
func resolveDevice(interactive bool) (simDevice, error) {
	devices, err := listSimulators()
	if err != nil {
		return simDevice{}, err
	}

	if len(devices) == 0 {
		return simDevice{}, fmt.Errorf("no iOS simulators found. Install them via Xcode → Settings → Platforms")
	}

	if !interactive {
		return pickDefaultDevice(devices), nil
	}

	return pickDeviceInteractively(devices)
}

func listSimulators() ([]simDevice, error) {
	out, err := exec.Command("xcrun", "simctl", "list", "devices", "available", "--json").Output()
	if err != nil {
		return nil, fmt.Errorf("listing simulators: %w", err)
	}

	var result struct {
		Devices map[string][]simDevice `json:"devices"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing simulator list: %w", err)
	}

	// Collect runtimes and sort by version (latest first)
	var runtimes []string
	for runtime := range result.Devices {
		runtimes = append(runtimes, runtime)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(runtimes)))

	var all []simDevice
	for _, runtime := range runtimes {
		devs := result.Devices[runtime]
		runtimeName := parseRuntimeName(runtime)
		if !strings.HasPrefix(runtimeName, "iOS") {
			continue
		}
		for _, d := range devs {
			if !d.Available {
				continue
			}
			d.Runtime = runtimeName
			all = append(all, d)
		}
	}

	return all, nil
}

// pickDefaultDevice finds the best default: prefer booted, then latest iPhone Pro.
// Devices are already sorted latest-runtime-first from listSimulators.
func pickDefaultDevice(devices []simDevice) simDevice {
	// Prefer already booted device
	for _, d := range devices {
		if d.State == "Booted" {
			return d
		}
	}

	// Prefer iPhone Pro (latest runtime is first in list)
	for _, d := range devices {
		if strings.Contains(d.Name, "iPhone") && strings.Contains(d.Name, "Pro") {
			return d
		}
	}
	for _, d := range devices {
		if strings.Contains(d.Name, "iPhone") {
			return d
		}
	}

	// Fallback to first available
	return devices[0]
}

func pickDeviceInteractively(devices []simDevice) (simDevice, error) {
	fmt.Println("Available iOS Simulators:")
	fmt.Println()

	for i, d := range devices {
		state := ""
		if d.State == "Booted" {
			state = " (booted)"
		}
		fmt.Printf("  %d) %s — %s%s\n", i+1, d.Name, d.Runtime, state)
	}

	fmt.Println()
	fmt.Print("Choose a device (number): ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return simDevice{}, fmt.Errorf("no input")
	}

	input := strings.TrimSpace(scanner.Text())
	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil || choice < 1 || choice > len(devices) {
		return simDevice{}, fmt.Errorf("invalid choice: %s", input)
	}

	return devices[choice-1], nil
}

// parseRuntimeName converts "com.apple.CoreSimulator.SimRuntime.iOS-26-2" → "iOS 26.2"
func parseRuntimeName(runtime string) string {
	// Remove prefix
	name := runtime
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	// "iOS-26-2" → "iOS 26.2"
	name = strings.Replace(name, "-", " ", 1) // first dash → space (after "iOS")
	name = strings.ReplaceAll(name, "-", ".")  // remaining dashes → dots
	return name
}

func hasFlag(flags ...string) bool {
	for _, arg := range os.Args[3:] {
		for _, f := range flags {
			if arg == f {
				return true
			}
		}
	}
	return false
}

func iosSimulatorSDK() string {
	out, err := exec.Command("xcrun", "--sdk", "iphonesimulator", "--show-sdk-path").Output()
	if err != nil {
		return "/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneSimulator.platform/Developer/SDKs/iPhoneSimulator.sdk"
	}
	return strings.TrimSpace(string(out))
}
