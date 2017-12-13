package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/blang/semver"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [OPTIONS]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		fmt.Fprintf(os.Stderr, "   -major, -minor, -patch, -pre   increase version part. default is -patch.\n")
		fmt.Fprintf(os.Stderr, "                                  only -pre may be combined with others.\n")
		fmt.Fprintf(os.Stderr, "   -version <version>             specify the release version. ignores other version modifiers.\n")
		fmt.Fprintf(os.Stderr, "   -pre-version <pre-release>     specify the pre-release version. implies -pre. default is 'RC' (when only -pre is set).\n")
		fmt.Fprintf(os.Stderr, "   -dry                           do not change anything. just print the result.\n")
		fmt.Fprintf(os.Stderr, "   -f                             ignore untracked & uncommitted changes.\n")
		fmt.Fprintf(os.Stderr, "   -h                             print this help.\n")
	}

	major := flag.Bool("major", false, "")
	minor := flag.Bool("minor", false, "")
	patch := flag.Bool("patch", false, "")
	pre := flag.Bool("pre", false, "")
	newVersion := flag.String("version", "", "")
	newPreVersion := flag.String("pre-version", "", "")
	dry := flag.Bool("dry", false, "")
	force := flag.Bool("f", false, "")

	flag.Parse()

	if newPreVersion != nil && *newPreVersion != "" {
		*pre = true
	}

	var err error
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
		if !*major && !*minor && !*patch && !*pre {
			*patch = true //Default is patch
		}
	}

	if *major {
		version.Major++
		version.Minor = 0
		version.Patch = 0
		version.Pre = nil
		version.Build = nil
	} else if *minor {
		version.Minor++
		version.Patch = 0
		version.Pre = nil
		version.Build = nil
	} else if *patch {
		version.Patch++
		version.Pre = nil
		version.Build = nil
	}

	if *pre {
		var preVersion semver.PRVersion

		if newPreVersion == nil || *newPreVersion == "" {
			if len(version.Pre) > 0 {
				preVersion = version.Pre[0]
			} else {
				preVersion, err = semver.NewPRVersion("RC0")
			}
			if preVersion.IsNum {
				preVersion.VersionNum++
			} else {
				bumpPreVersion(&preVersion)
			}
		} else {
			preVersion, err = semver.NewPRVersion(*newPreVersion)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid : %v\n", err)
				os.Exit(1)
			}
		}

		version.Pre = []semver.PRVersion{preVersion}
	}

	dryRunInfo := ""
	if *dry {
		dryRunInfo = "[dry-run] "
	}
	fmt.Fprintf(os.Stdout, "%vReleasing version %v.\n", dryRunInfo, version.String())

	err = checkRepoStatus(*force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "%vTagging.\n", dryRunInfo)
	if !*dry {
		if err = tag(version); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stdout, "%vPushing tag.\n", dryRunInfo)
	if !*dry {
		if err = push(); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stdout, "%vRelease %v successful.\n", dryRunInfo, version.String())

}

func getVersionFromGit() (semver.Version, error) {
	cmd := exec.Command("git", "tag", "--sort=-creatordate")
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
	fmt.Fprintf(os.Stdout, "Latest git version is '%v'.\n", versionText)
	return semver.Parse(versionText)
}

func checkRepoStatus(force bool) error {
	if !isGitAvailable() {
		return fmt.Errorf("git is not available.")
	}

	if !isGitRepository() {
		return fmt.Errorf("Not a git repository.")
	}

	if !force && hasChanges() {
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

	if err != nil || lineCount > 0 {
		return true
	}

	return false
}

func hasStagedChanges() bool {
	cmd := exec.Command("git", "cherry", "-v")
	lineCount, err := countOutputLines(cmd)

	if err != nil || lineCount > 0 {
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

	if err != nil {
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

func bumpPreVersion(version *semver.PRVersion) {
	re := regexp.MustCompile("[0-9]+$")
	loc := re.FindStringIndex(version.String())

	if len(loc) == 0 { //no matches
		version.VersionStr = version.VersionStr + "1"
		return
	}

	versionName := version.VersionStr[:loc[0]]
	versionNum, _ := strconv.Atoi(version.VersionStr[loc[0]:])

	version.VersionStr = fmt.Sprintf("%s%d", versionName, versionNum+1)

}
