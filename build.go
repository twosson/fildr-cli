package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var (
	VERSION    = "v0.0.1"
	GOPATH     = os.Getenv("GOPATH")
	GIT_COMMIT = gitCommit()
	BUILD_TIME = time.Now().UTC().Format(time.RFC3339)
	LD_FLAGS   = fmt.Sprintf("-X \"main.buildTime=%s\" -X main.gitCommit=%s", BUILD_TIME, GIT_COMMIT)
	GO_FLAGS   = fmt.Sprintf("-ldflags=%s", LD_FLAGS)
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "build.go",
		Short: "Build tools for fildr-cli",
	}

	rootCmd.Name()

	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "go-install",
			Short: "install build tools",
			Run: func(cmd *cobra.Command, args []string) {
				goInstall()
			},
		},
		&cobra.Command{
			Use:   "generate",
			Short: "update generated artifacts",
			Run: func(cmd *cobra.Command, args []string) {
				generate()
			},
		},
		&cobra.Command{
			Use:   "version",
			Short: "version",
			Run: func(cmd *cobra.Command, args []string) {
				version()
			},
		},
		&cobra.Command{
			Use:   "vet",
			Short: "lint server code",
			Run: func(cmd *cobra.Command, args []string) {
				vet()
			},
		},
		&cobra.Command{
			Use:   "test",
			Short: "run server tests",
			Run: func(cmd *cobra.Command, args []string) {
				test()
			},
		},
		&cobra.Command{
			Use:   "build",
			Short: "server build, skipping tests",
			Run: func(cmd *cobra.Command, args []string) {
				build()
			},
		},
		&cobra.Command{
			Use:   "release",
			Short: "tag and push a release",
			Run: func(cmd *cobra.Command, args []string) {
				release()
			},
		},
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runCmd(command string, env map[string]string, args ...string) {
	cmd := newCmd(command, env, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	log.Printf("Running: %s\n", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func runCmdIn(dir, command string, env map[string]string, args ...string) {
	cmd := newCmd(command, env, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	log.Printf("Running in %s: %s\n", dir, cmd.String())
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func newCmd(command string, env map[string]string, args ...string) *exec.Cmd {
	realCommand, err := exec.LookPath(command)
	if err != nil {
		log.Fatalf("unable to find command '%s'", command)
	}

	cmd := exec.Command(realCommand, args...)
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	return cmd
}

func goInstall() {
	pkgs := []string{
		"github.com/GeertJohan/go.rice",
		"github.com/GeertJohan/go.rice/rice",
		"github.com/golang/mock/gomock",
		"github.com/golang/mock/mockgen",
		"github.com/golang/protobuf/protoc-gen-go",
	}
	for _, pkg := range pkgs {
		runCmd("go", map[string]string{"GO111MODULE": "on"}, "install", pkg)
	}
}

func generate() {
	removeFakes()
	runCmd("go", nil, "generate", "-v", "./pkg/...", "./internal/...")
}

func version() {
	fmt.Println(VERSION)
}

func gitCommit() string {
	cmd := newCmd("git", nil, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("gitCommit: %s", err)
		return ""
	}
	return fmt.Sprintf("%s", out)
}

func release() {
	runCmd("git", nil, "tag", "-a", VERSION, "-m", fmt.Sprintf("\"Release %s\"", VERSION))
	runCmd("git", nil, "push", "--follow-tags")
}

func vet() {
	runCmd("go", nil, "vet", "./internal/...", "./pkg/...")
}

func test() {
	runCmd("go", nil, "test", "-v", "./internal/...", "./pkg/...")
}

func build() {
	newPath := filepath.Join(".", "build")
	os.MkdirAll(newPath, 0755)

	artifact := "octant"
	if runtime.GOOS == "windows" {
		artifact = "octant.exe"
	}
	runCmd("go", nil, "build", "-o", "build/"+artifact, GO_FLAGS, "-v", "./cmd/octant")
}

func removeFakes() {
	checkDirs := []string{"pkg", "internal"}
	fakePaths := []string{}

	for _, dir := range checkDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}
			if info.Name() == "fake" {
				fakePaths = append(fakePaths, filepath.Join(path, info.Name()))
			}
			return nil
		})
		if err != nil {
			log.Fatalf("generate (%s): %s", dir, err)
		}
	}

	log.Print("Removing fakes from pkg/ and internal/")
	for _, p := range fakePaths {
		os.RemoveAll(p)
	}
}
