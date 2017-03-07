package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/blang/semver"
	"bytes"
	"os/exec"
	"io/ioutil"
	"bufio"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [OPTIONS]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		fmt.Fprintf(os.Stderr, "   -major, -minor, -patch   increase version part. default is -patch\n")
		fmt.Fprintf(os.Stderr, "   -build <build-name>      include additional build name (e.g. alpha)\n")
		fmt.Fprintf(os.Stderr, "   -version <version>       specify the release version. ignores other version modifiers.\n")
		fmt.Fprintf(os.Stderr, "   -h                       print this help.\n")
	}

	major := flag.Bool("major", false, "increase major version")
	minor := flag.Bool("minor", false, "increase minor version")
	patch := flag.Bool("patch", false, "increase patch version")
	build := flag.String("build", "", "set build")
	newVersion := flag.String("version", "", "set build")

	flag.Parse()

	err := checkRepoStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	version := semver.Version{}
	if newVersion != nil && *newVersion != "" {
		ver, err := semver.New(*newVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Version string '%v' is not a valid version.\n", newVersion)
			os.Exit(1)
		}
		version = *ver
	} else {
		fmt.Fprintf(os.Stdout, "Retrieving old version from git.\n")

		version, err = getVersionFromGit()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get version from git: %v\n", err)
			os.Exit(1)
		}
		if (major == nil || *major == false) && (minor == nil || *minor == false) && (patch == nil || *patch == false) {
			*patch = true
		}
	}

	if *major {
		version.Major++
	}
	if *minor {
		version.Minor++
	}
	if *patch {
		version.Patch++
	}
	if *build != "" {
		version.Build = []string{*build}
	}

	fmt.Fprintf(os.Stdout, "Tagging version %v.\n", version.String())
	if err = tag(version); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Pushing tag.\n")
	if err = push(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Release %v successful.\n", version.String())

}

func getVersionFromGit() (semver.Version, error) {
	cmd := exec.Command("git", "tag", "--sort=-version:refname")
	result := &bytes.Buffer{}
	cmd.Stdout = result
	cmd.Stderr = result
	if err := cmd.Run(); err != nil {
		return semver.Version{}, fmt.Errorf("Cannot read latest version output from git. %v.\nGit message: %v", err, result.String())
	}

	scanner := bufio.NewScanner(result)
	if !scanner.Scan() {
		return semver.Version{}, fmt.Errorf("No git versions found.")

	}

	versionText := scanner.Text()
	fmt.Fprintf(os.Stdout, "Latest git version is '%v'.", versionText)
	return semver.Parse(versionText)
}

func checkRepoStatus() error {
	if !isGitAvailable() {
		return fmt.Errorf("git is not available.")
	}

	if !isGitRepository() {
		return fmt.Errorf("Not a git repository.")
	}

	if hasChanges() {
		return fmt.Errorf("Your repository has uncommited changes.")
	}

	if hasStagedChanges() {
		return fmt.Errorf("Your repository has unpushed changes. Please push.")
	}

	if val, err := isBehind(); err != nil {
		return fmt.Errorf("Could not determine remote status: %v.", err)
	} else if val == true {
		return fmt.Errorf("Your branch is behind the remote. Please pull.")
	}

	return nil
}

func isGitAvailable() bool {
	cmd := exec.Command("which", "git")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}
func isGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func hasChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	lineCount, err := countOutputLines(cmd)

	if err!= nil || lineCount > 0 {
		return true
	}

	return false
}


func hasStagedChanges() bool {
	cmd := exec.Command("git", "cherry", "-v")
	lineCount, err := countOutputLines(cmd)

	if err!= nil || lineCount > 0 {
		return true
	}

	return false
}

func isBehind() (bool, error) {
	cmd := exec.Command("git", "fetch")
	result := &bytes.Buffer{}
	cmd.Stdout = result
	cmd.Stderr = result
	if err := cmd.Run(); err != nil {
		return true, fmt.Errorf("Failed to fetch remote. %v", err)
	}

	cmd = exec.Command("git", "--no-pager", "log", "HEAD..@{u}", "--oneline")
	lineCount, err := countOutputLines(cmd)

	if err!= nil {
		return true, err
	}

	if lineCount > 0 {
		return true, nil
	}

	return false, nil
}

func countOutputLines(cmd *exec.Cmd) (int, error) {
	result := &bytes.Buffer{}
	cmd.Stdout = result
	if err := cmd.Run(); err != nil {
		return 1, err
	}

	scanner := bufio.NewScanner(result)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	return lineCount, nil
}

func tag(version semver.Version) error {
	cmd := exec.Command("git", "tag", version.String())
	result := &bytes.Buffer{}
	cmd.Stdout = result
	cmd.Stderr = result
	if err := cmd.Run(); err != nil {
		output, err := ioutil.ReadAll(result)
		return fmt.Errorf("Failed to tag. %v\n%v", err, string(output))
	}

	return nil
}

func push() error {
	cmd := exec.Command("git", "push", "--tag")
	result := &bytes.Buffer{}
	cmd.Stdout = result
	cmd.Stderr = result
	if err := cmd.Run(); err != nil {
		output, err := ioutil.ReadAll(result)
		return fmt.Errorf("Failed to push tags. %v\n%v", err, string(output))
	}

	return nil
}