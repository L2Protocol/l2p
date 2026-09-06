// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

//go:build none
// +build none

/*
The ci command is called from Continuous Integration scripts.

Usage: go run build/ci.go <command> <command flags/arguments>

Available commands are:

	lint           -- runs certain pre-selected linters
	check_generate -- verifies that 'go generate' and 'go mod tidy' do not produce changes
	check_baddeps  -- verifies that certain dependencies are avoided

	install    [ -arch architecture ] [ -cc compiler ] [ packages... ] -- builds packages and executables
	test       [ -coverage ] [ packages... ]                           -- runs the tests

	archive    [ -arch architecture ] [ -type zip|tar ] [ -signify key-envvar ] [ -upload dest ] -- archives build artifacts

For all commands, -n prevents execution of external programs (dry run mode).
*/
package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/signify"
	"github.com/ethereum/go-ethereum/internal/build"
	"github.com/ethereum/go-ethereum/internal/download"
	"github.com/ethereum/go-ethereum/internal/version"
)

var (
	// Files that end up in the geth*.zip archive.
	gethArchiveFiles = []string{
		"COPYING",
		executablePath("geth"),
	}

	// Files that end up in the geth-alltools*.zip archive.
	allToolsArchiveFiles = []string{
		"COPYING",
		executablePath("abigen"),
		executablePath("bootnode"),
		executablePath("evm"),
		executablePath("geth"),
		executablePath("rlpdump"),
		executablePath("clef"),
	}

	// This is where the tests should be unpacked.
	executionSpecTestsDir = "tests/spec-tests"
)

var GOBIN, _ = filepath.Abs(filepath.Join("build", "bin"))

func executablePath(name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(GOBIN, name)
}

func main() {
	log.SetFlags(log.Lshortfile)

	if !common.FileExist(filepath.Join("build", "ci.go")) {
		log.Fatal("this script must be run from the root of the repository")
	}
	if len(os.Args) < 2 {
		log.Fatal("need subcommand as first argument")
	}
	switch os.Args[1] {
	case "install":
		doInstall(os.Args[2:])
	case "test":
		doTest(os.Args[2:])
	case "lint":
		doLint(os.Args[2:])
	case "check_generate":
		doCheckGenerate()
	case "check_baddeps":
		doCheckBadDeps()
	case "archive":
		doArchive(os.Args[2:])
	case "dockerx":
		doDockerBuildx(os.Args[2:])
	case "sanitycheck":
		doSanityCheck()
	default:
		log.Fatal("unknown command ", os.Args[1])
	}
}

// Compiling

func doInstall(cmdline []string) {
	var (
		dlgo       = flag.Bool("dlgo", false, "Download Go and build with it")
		arch       = flag.String("arch", "", "Architecture to cross build for")
		cc         = flag.String("cc", "", "C compiler to cross build with")
		staticlink = flag.Bool("static", false, "Create statically-linked executable")
		output     = flag.String("o", "", "Output directory for build artifacts")
	)
	flag.CommandLine.Parse(cmdline)
	env := build.Env()

	// Configure the toolchain.
	tc := build.GoToolchain{GOARCH: *arch, CC: *cc}
	if *dlgo {
		csdb := download.MustLoadChecksums("build/checksums.txt")
		tc.Root = build.DownloadGo(csdb)
	}
	// Disable CLI markdown doc generation in release builds.
	buildTags := []string{"urfave_cli_no_docs"}

	// Enable linking the CKZG library since we can make it work with additional flags.
	if env.UbuntuVersion != "trusty" {
		buildTags = append(buildTags, "ckzg")
	}

	// Configure the build.
	gobuild := tc.Go("build", buildFlags(env, *staticlink, buildTags)...)

	// We use -trimpath to avoid leaking local paths into the built executables.
	gobuild.Args = append(gobuild.Args, "-trimpath")

	// Show packages during build.
	gobuild.Args = append(gobuild.Args, "-v")

	// Now we choose what we're even building.
	// Default: collect all 'main' packages in cmd/ and build those.
	packages := flag.Args()
	if len(packages) == 0 {
		packages = build.FindMainPackages(&tc, "./cmd/...")
	}

	// Do the build!
	for _, pkg := range packages {
		args := slices.Clone(gobuild.Args)
		outputPath := executablePath(path.Base(pkg))
		if output != nil && *output != "" {
			outputPath = *output
		}
		args = append(args, "-o", outputPath)
		args = append(args, pkg)
		build.MustRun(&exec.Cmd{Path: gobuild.Path, Args: args, Env: gobuild.Env})
	}
}

// buildFlags returns the go tool flags for building.
func buildFlags(env build.Environment, staticLinking bool, buildTags []string) (flags []string) {
	var ld []string
	// See https://github.com/golang/go/issues/33772#issuecomment-528176001
	// We need to set --buildid to the linker here, and also pass --build-id to the
	// cgo-linker further down.
	ld = append(ld, "--buildid=none")
	if env.Commit != "" {
		ld = append(ld, "-X", "github.com/ethereum/go-ethereum/internal/version.gitCommit="+env.Commit)
		ld = append(ld, "-X", "github.com/ethereum/go-ethereum/internal/version.gitDate="+env.Date)
	}
	// Strip DWARF on darwin. This used to be required for certain things,
	// and there is no downside to this, so we just keep doing it.
	if runtime.GOOS == "darwin" {
		ld = append(ld, "-s")
	}
	if runtime.GOOS == "linux" {
		// Enforce the stacksize to 8M, which is the case on most platforms apart from
		// alpine Linux.
		// See https://sourceware.org/binutils/docs-2.23.1/ld/Options.html#Options
		// regarding the options --build-id=none and --strip-all. It is needed for
		// reproducible builds; removing references to temporary files in C-land, and
		// making build-id reproducibly absent.
		extld := []string{"-Wl,-z,stack-size=0x800000,--build-id=none,--strip-all"}
		if staticLinking {
			extld = append(extld, "-static")
			// Under static linking, use of certain glibc features must be
			// disabled to avoid shared library dependencies.
			buildTags = append(buildTags, "osusergo", "netgo")
		}
		ld = append(ld, "-extldflags", "'"+strings.Join(extld, " ")+"'")
	}
	if len(ld) > 0 {
		flags = append(flags, "-ldflags", strings.Join(ld, " "))
	}
	if len(buildTags) > 0 {
		flags = append(flags, "-tags", strings.Join(buildTags, ","))
	}
	return flags
}

// Running The Tests
//
// "tests" also includes static analysis tools such as vet.

func doTest(cmdline []string) {
	var (
		dlgo     = flag.Bool("dlgo", false, "Download Go and build with it")
		arch     = flag.String("arch", "", "Run tests for given architecture")
		cc       = flag.String("cc", "", "Sets C compiler binary")
		coverage = flag.Bool("coverage", false, "Whether to record code coverage")
		verbose  = flag.Bool("v", false, "Whether to log verbosely")
		timeout  = flag.String("timeout", "10m", `Timeout of running tests`)
		race     = flag.Bool("race", false, "Execute the race detector")
		short    = flag.Bool("short", false, "Pass the 'short'-flag to go test")
		// cachedir = flag.String("cachedir", "./build/cache", "directory for caching downloads")
	)
	flag.CommandLine.Parse(cmdline)

	// Get test fixtures.
	csdb := download.MustLoadChecksums("build/checksums.txt")
	// downloadSpecTestFixtures(csdb, *cachedir)

	// Configure the toolchain.
	tc := build.GoToolchain{GOARCH: *arch, CC: *cc}
	if *dlgo {
		tc.Root = build.DownloadGo(csdb)
	}
	gotest := tc.Go("test")

	// CI needs a bit more time for the statetests (default 45m).
	gotest.Args = append(gotest.Args, "-timeout=45m")

	// Enable CKZG backend in CI.
	gotest.Args = append(gotest.Args, "-tags=ckzg")

	// Enable integration-tests
	gotest.Args = append(gotest.Args, "-tags=integrationtests")

	// Test a single package at a time. CI builders are slow
	// and some tests run into timeouts under load.
	gotest.Args = append(gotest.Args, "-p", "1")
	if *coverage {
		gotest.Args = append(gotest.Args, "-covermode=atomic", "-cover")
	}
	if *verbose {
		gotest.Args = append(gotest.Args, "-v")
	}
	if *timeout != "" {
		gotest.Args = append(gotest.Args, []string{"-timeout", *timeout}...)
	}
	if *race {
		gotest.Args = append(gotest.Args, "-race")
	}
	if *short {
		gotest.Args = append(gotest.Args, "-short")
	}

	packages := []string{"./..."}
	if len(flag.CommandLine.Args()) > 0 {
		packages = flag.CommandLine.Args()
	}
	gotest.Args = append(gotest.Args, packages...)
	build.MustRun(gotest)
}

// downloadSpecTestFixtures downloads and extracts the execution-spec-tests fixtures.
func downloadSpecTestFixtures(csdb *download.ChecksumDB, cachedir string) string {
	ext := ".tar.gz"
	base := "fixtures_develop"
	archivePath := filepath.Join(cachedir, base+ext)
	if err := csdb.DownloadFileFromKnownURL(archivePath); err != nil {
		log.Fatal(err)
	}
	if err := build.ExtractArchive(archivePath, executionSpecTestsDir); err != nil {
		log.Fatal(err)
	}
	return filepath.Join(cachedir, base)
}

// doCheckTidy assets that the Go modules files are tidied already.
func doCheckTidy() {
}

// doCheckGenerate ensures that re-generating generated files does not cause
// any mutations in the source file tree.
func doCheckGenerate() {
	var (
		cachedir = flag.String("cachedir", "./build/cache", "directory for caching binaries.")
		tc       = new(build.GoToolchain)
	)
	// Compute the origin hashes of all the files
	var hashes map[string][32]byte

	var err error
	hashes, err = build.HashFolder(".", []string{"tests/testdata", "build/cache", ".git"})
	if err != nil {
		log.Fatal("Error computing hashes", "err", err)
	}
	// Run any go generate steps we might be missing
	var (
		protocPath      = downloadProtoc(*cachedir)
		protocGenGoPath = downloadProtocGenGo(*cachedir)
	)
	c := tc.Go("generate", "./...")
	pathList := []string{filepath.Join(protocPath, "bin"), protocGenGoPath, os.Getenv("PATH")}
	c.Env = append(c.Env, "PATH="+strings.Join(pathList, string(os.PathListSeparator)))
	build.MustRun(c)

	// Check if generate file hashes have changed
	generated, err := build.HashFolder(".", []string{"tests/testdata", "build/cache", ".git"})
	if err != nil {
		log.Fatalf("Error re-computing hashes: %v", err)
	}
	updates := build.DiffHashes(hashes, generated)
	for _, file := range updates {
		log.Printf("File changed: %s", file)
	}
	if len(updates) != 0 {
		log.Fatal("One or more generated files were updated by running 'go generate ./...'")
	}
	fmt.Println("No stale files detected.")

	// Run go mod tidy check.
	build.MustRun(tc.Go("mod", "tidy", "-diff"))
	fmt.Println("No untidy module files detected.")
}

// doCheckBadDeps verifies whether certain unintended dependencies between some
// packages leak into the codebase due to a refactor. This is not an exhaustive
// list, rather something we build up over time at sensitive places.
func doCheckBadDeps() {
	baddeps := [][2]string{
		// Rawdb tends to be a dumping ground for db utils, sometimes leaking the db itself
		{"github.com/ethereum/go-ethereum/core/rawdb", "github.com/ethereum/go-ethereum/ethdb/leveldb"},
		{"github.com/ethereum/go-ethereum/core/rawdb", "github.com/ethereum/go-ethereum/ethdb/pebbledb"},
	}
	tc := new(build.GoToolchain)

	var failed bool
	for _, rule := range baddeps {
		out, err := tc.Go("list", "-deps", rule[0]).CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to list '%s' dependencies: %v", rule[0], err)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(line) == rule[1] {
				log.Printf("Found bad dependency '%s' -> '%s'", rule[0], rule[1])
				failed = true
			}
		}
	}
	if failed {
		log.Fatalf("Bad dependencies detected.")
	}
	fmt.Println("No bad dependencies detected.")
}

// doLint runs golangci-lint on requested packages.
func doLint(cmdline []string) {
	var (
		cachedir = flag.String("cachedir", "./build/cache", "directory for caching golangci-lint binary.")
	)
	flag.CommandLine.Parse(cmdline)
	packages := []string{"./..."}
	if len(flag.CommandLine.Args()) > 0 {
		packages = flag.CommandLine.Args()
	}

	linter := downloadLinter(*cachedir)
	lflags := []string{"run", "--config", ".golangci.yml"}
	build.MustRunCommandWithOutput(linter, append(lflags, packages...)...)
	fmt.Println("You have achieved perfection.")
}

// downloadLinter downloads and unpacks golangci-lint.
func downloadLinter(cachedir string) string {
	csdb := download.MustLoadChecksums("build/checksums.txt")
	version, err := csdb.FindVersion("golangci")
	if err != nil {
		log.Fatal(err)
	}
	arch := runtime.GOARCH
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	if arch == "arm" {
		arch += "v" + os.Getenv("GOARM")
	}
	base := fmt.Sprintf("golangci-lint-%s-%s-%s", version, runtime.GOOS, arch)
	archivePath := filepath.Join(cachedir, base+ext)
	if err := csdb.DownloadFileFromKnownURL(archivePath); err != nil {
		log.Fatal(err)
	}
	if err := build.ExtractArchive(archivePath, cachedir); err != nil {
		log.Fatal(err)
	}
	return filepath.Join(cachedir, base, "golangci-lint")
}

// protocArchiveBaseName returns the name of the protoc archive file for
// the current system, stripped of version and file suffix.
func protocArchiveBaseName() (string, error) {
	switch runtime.GOOS + "-" + runtime.GOARCH {
	case "windows-amd64":
		return "win64", nil
	case "windows-386":
		return "win32", nil
	case "linux-arm64":
		return "linux-aarch_64", nil
	case "linux-386":
		return "linux-x86_32", nil
	case "linux-amd64":
		return "linux-x86_64", nil
	case "darwin-arm64":
		return "osx-aarch_64", nil
	case "darwin-amd64":
		return "osx-x86_64", nil
	default:
		return "", fmt.Errorf("no prebuilt release of protoc available for this system (os: %s, arch: %s)", runtime.GOOS, runtime.GOARCH)
	}
}

// downloadProtocGenGo downloads protoc-gen-go, which is used by protoc
// in the generate command.  It returns the full path of the directory
// containing the 'protoc-gen-go' executable.
func downloadProtocGenGo(cachedir string) string {
	csdb := download.MustLoadChecksums("build/checksums.txt")
	version, err := csdb.FindVersion("protoc-gen-go")
	if err != nil {
		log.Fatal(err)
	}
	baseName := fmt.Sprintf("protoc-gen-go.v%s.%s.%s", version, runtime.GOOS, runtime.GOARCH)
	archiveName := baseName
	if runtime.GOOS == "windows" {
		archiveName += ".zip"
	} else {
		archiveName += ".tar.gz"
	}

	archivePath := path.Join(cachedir, archiveName)
	if err := csdb.DownloadFileFromKnownURL(archivePath); err != nil {
		log.Fatal(err)
	}
	extractDest := filepath.Join(cachedir, baseName)
	if err := build.ExtractArchive(archivePath, extractDest); err != nil {
		log.Fatal(err)
	}
	extractDest, err = filepath.Abs(extractDest)
	if err != nil {
		log.Fatal("error resolving absolute path for protoc", "err", err)
	}
	return extractDest
}

// downloadProtoc downloads the prebuilt protoc binary used to lint generated
// files as a CI step.  It returns the full path to the directory containing
// the protoc executable.
func downloadProtoc(cachedir string) string {
	csdb := download.MustLoadChecksums("build/checksums.txt")
	version, err := csdb.FindVersion("protoc")
	if err != nil {
		log.Fatal(err)
	}
	baseName, err := protocArchiveBaseName()
	if err != nil {
		log.Fatal(err)
	}

	fileName := fmt.Sprintf("protoc-%s-%s", version, baseName)
	archiveFileName := fileName + ".zip"
	archivePath := filepath.Join(cachedir, archiveFileName)
	if err := csdb.DownloadFileFromKnownURL(archivePath); err != nil {
		log.Fatal(err)
	}
	extractDest := filepath.Join(cachedir, fileName)
	if err := build.ExtractArchive(archivePath, extractDest); err != nil {
		log.Fatal(err)
	}
	extractDest, err = filepath.Abs(extractDest)
	if err != nil {
		log.Fatal("error resolving absolute path for protoc", "err", err)
	}
	return extractDest
}

// Release Packaging
func doArchive(cmdline []string) {
	var (
		arch    = flag.String("arch", runtime.GOARCH, "Architecture cross packaging")
		atype   = flag.String("type", "zip", "Type of archive to write (zip|tar)")
		signify = flag.String("signify", "", `Environment variable holding the signify key (e.g. LINUX_SIGNIFY_KEY)`)
		upload  = flag.String("upload", "", `Destination to upload the archives (usually "gethstore/builds")`)
		ext     string
	)
	flag.CommandLine.Parse(cmdline)
	switch *atype {
	case "zip":
		ext = ".zip"
	case "tar":
		ext = ".tar.gz"
	default:
		log.Fatal("unknown archive type: ", atype)
	}

	var (
		env      = build.Env()
		basegeth = archiveBasename(*arch, version.Archive(env.Commit))
		geth     = "geth-" + basegeth + ext
		alltools = "geth-alltools-" + basegeth + ext
	)
	maybeSkipArchive(env)
	if err := build.WriteArchive(geth, gethArchiveFiles); err != nil {
		log.Fatal(err)
	}
	if err := build.WriteArchive(alltools, allToolsArchiveFiles); err != nil {
		log.Fatal(err)
	}
	for _, archive := range []string{geth, alltools} {
		if err := archiveUpload(archive, *upload, *signify); err != nil {
			log.Fatal(err)
		}
	}
}

func archiveBasename(arch string, archiveVersion string) string {
	platform := runtime.GOOS + "-" + arch
	if arch == "arm" {
		platform += os.Getenv("GOARM")
	}
	if arch == "android" {
		platform = "android-all"
	}
	if arch == "ios" {
		platform = "ios-all"
	}
	return platform + "-" + archiveVersion
}

func archiveUpload(archive string, blobstore string, signifyVar string) error {
	// If signing was requested, generate the signature files
	if signifyVar != "" {
		key := os.Getenv(signifyVar)
		untrustedComment := "verify with geth-release.pub"
		trustedComment := fmt.Sprintf("%s (%s)", archive, time.Now().UTC().Format(time.RFC1123))
		if err := signify.SignFile(archive, archive+".sig", key, untrustedComment, trustedComment); err != nil {
			return err
		}
	}
	// If uploading to Azure was requested, push the archive possibly with its signature
	if blobstore != "" {
		auth := build.AzureBlobstoreConfig{
			Account:   strings.Split(blobstore, "/")[0],
			Token:     os.Getenv("AZURE_BLOBSTORE_TOKEN"),
			Container: strings.SplitN(blobstore, "/", 2)[1],
		}
		if err := build.AzureBlobstoreUpload(archive, filepath.Base(archive), auth); err != nil {
			return err
		}
		if signifyVar != "" {
			if err := build.AzureBlobstoreUpload(archive+".sig", filepath.Base(archive+".sig"), auth); err != nil {
				return err
			}
		}
	}
	return nil
}

// skips archiving for some build configurations.
func maybeSkipArchive(env build.Environment) {
	if env.IsPullRequest {
		log.Printf("skipping archive creation because this is a PR build")
		os.Exit(0)
	}
	if env.Branch != "master" && !strings.HasPrefix(env.Tag, "v1.") {
		log.Printf("skipping archive creation because branch %q, tag %q is not on the inclusion list", env.Branch, env.Tag)
		os.Exit(0)
	}
}

// Builds the docker images and optionally uploads them to Docker Hub.
func doDockerBuildx(cmdline []string) {
	var (
		platform = flag.String("platform", "", `Push a multi-arch docker image for the specified architectures (usually "linux/amd64,linux/arm64")`)
		hubImage = flag.String("hub", "ethereum/client-go", `Where to upload the docker image`)
		upload   = flag.Bool("upload", false, `Whether to trigger upload`)
	)
	flag.CommandLine.Parse(cmdline)

	// Skip building and pushing docker images for PR builds
	env := build.Env()
	maybeSkipArchive(env)

	// Retrieve the upload credentials and authenticate
	user := getenvBase64("DOCKER_HUB_USERNAME")
	pass := getenvBase64("DOCKER_HUB_PASSWORD")

	if len(user) > 0 && len(pass) > 0 {
		auther := exec.Command("docker", "login", "-u", string(user), "--password-stdin")
		auther.Stdin = bytes.NewReader(pass)
		build.MustRun(auther)
	}
	// Retrieve the version infos to build and push to the following paths:
	//  - ethereum/client-go:latest                            - Pushes to the master branch, Geth only
	//  - ethereum/client-go:stable                            - Version tag publish on GitHub, Geth only
	//  - ethereum/client-go:alltools-latest                   - Pushes to the master branch, Geth & tools
	//  - ethereum/client-go:alltools-stable                   - Version tag publish on GitHub, Geth & tools
	//  - ethereum/client-go:release-<major>.<minor>           - Version tag publish on GitHub, Geth only
	//  - ethereum/client-go:alltools-release-<major>.<minor>  - Version tag publish on GitHub, Geth & tools
	//  - ethereum/client-go:v<major>.<minor>.<patch>          - Version tag publish on GitHub, Geth only
	//  - ethereum/client-go:alltools-v<major>.<minor>.<patch> - Version tag publish on GitHub, Geth & tools
	var tags []string

	switch {
	case env.Branch == "master":
		tags = []string{"latest"}
	case strings.HasPrefix(env.Tag, "v1."):
		tags = []string{"stable", fmt.Sprintf("release-%v", version.Family), "v" + version.Semantic}
	}
	// Need to create a mult-arch builder
	check := exec.Command("docker", "buildx", "inspect", "multi-arch-builder")
	if check.Run() != nil {
		build.MustRunCommand("docker", "buildx", "create", "--use", "--name", "multi-arch-builder", "--platform", *platform)
	}

	for _, spec := range []struct {
		file string
		base string
	}{
		{file: "Dockerfile", base: fmt.Sprintf("%s:", *hubImage)},
		{file: "Dockerfile.alltools", base: fmt.Sprintf("%s:alltools-", *hubImage)},
	} {
		for _, tag := range tags { // latest, stable etc
			gethImage := fmt.Sprintf("%s%s", spec.base, tag)
			cmd := exec.Command("docker", "buildx", "build",
				"--build-arg", "COMMIT="+env.Commit,
				"--build-arg", "VERSION="+version.WithMeta,
				"--build-arg", "BUILDNUM="+env.Buildnum,
				"--tag", gethImage,
				"--platform", *platform,
				"--file", spec.file,
			)
			if *upload {
				cmd.Args = append(cmd.Args, "--push")
			}
			cmd.Args = append(cmd.Args, ".")
			build.MustRun(cmd)
		}
	}
}

// Debian Packaging
func getenvBase64(variable string) []byte {
	dec, err := base64.StdEncoding.DecodeString(os.Getenv(variable))
	if err != nil {
		log.Fatal("invalid base64 " + variable)
	}
	return []byte(dec)
}

func doSanityCheck() {
	csdb := download.MustLoadChecksums("build/checksums.txt")
	csdb.DownloadAndVerifyAll()
}
