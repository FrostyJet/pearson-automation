// Example usage:
// 	go run cmd/release/release.go -from release/Global-PMC-R35.3a -to release/Global-PMC-R35.3b

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

func failOnError(msg string, err error) {
	if err != nil {
		log.Panicf("%s : %s", msg, err)
	}
}

var commands map[string]string = map[string]string{
	"git": "/usr/bin/git",
}

const repoBaseDir = "/Users/eduard/workspace/pearson"

func main() {
	fromBranch := flag.String("from", "", "Environment to deploy")
	newBranch := flag.String("to", "", "Branch name to deploy")

	flag.Parse()

	if *fromBranch == "" || *newBranch == "" {
		log.Fatalf("please specify new release branch (-to) name and the starting branch name (-from)")
	}
	fmt.Printf("Starting new release: '%s' -> '%s'\n", *fromBranch, *newBranch)

	repoNames := []string{"pmc-react-browse", "pmc-react-login", "pmc-react-myaccount", "pmc-react-replocator", "pmc-react-shared"}

	for _, repoName := range repoNames {
		log.Printf("working on repository %s\n", repoName)

		// reset from current branch
		branchName := currentBranchName(repoName)
		resetLocalChanges(repoName)
		log.Printf("[%s] checked out from branch: %s\n", repoName, branchName)

		// checkout to "from" branch
		checkoutToBranch(repoName, *fromBranch)
		log.Printf("[%s] switched to branch: %s\n", repoName, branchName)

		// pull latest versions
		pullLatestChanges(repoName)
		log.Printf("[%s] pulled latest changes from origin\n", repoName)

		// create new branch
		createNewBranch(repoName, *newBranch)
		log.Printf("[%s] created new branch %s\n", repoName, *newBranch)

		// re-point submodule
		if repoName != "pmc-react-shared" {
			setSubmoduleTo(repoName, *newBranch)
			log.Printf("[%s] updated .gitmodules file to use %s branch from shared\n", repoName, *newBranch)
		}

		// push changes to remote
		pushChangesToRemote(repoName, *newBranch, fmt.Sprintf("'chore: create new release: %s'", *newBranch))
		log.Printf("[%s] done\n", repoName)
	}
}

func resetLocalChanges(repo string) string {
	cmd := exec.Command(commands["git"], "checkout", ".")
	cmd.Dir = fmt.Sprintf("%s/%s", repoBaseDir, repo)

	stdout, err := cmd.Output()
	failOnError("could not reset local changes", err)

	return string(stdout)
}

func currentBranchName(repo string) string {
	cmd := exec.Command(commands["git"], "branch", "--show-current")
	cmd.Dir = fmt.Sprintf("%s/%s", repoBaseDir, repo)

	stdout, err := cmd.Output()
	failOnError("could print current branch name", err)

	return string(stdout)
}

func checkoutToBranch(repo, branch string) string {
	cmd := exec.Command(commands["git"], "checkout", branch)
	cmd.Dir = fmt.Sprintf("%s/%s", repoBaseDir, repo)

	stdout, err := cmd.Output()
	failOnError(fmt.Sprintf("could checkout to branch %s\n", repo), err)

	return string(stdout)
}

func pullLatestChanges(repo string) string {
	cmd := exec.Command(commands["git"], "pull", "--rebase")
	cmd.Dir = fmt.Sprintf("%s/%s", repoBaseDir, repo)

	stdout, err := cmd.Output()
	failOnError("could not pull latest version", err)

	return string(stdout)
}

func createNewBranch(repo string, branch string) string {
	cmd := exec.Command(commands["git"], "checkout", "-b", branch)
	cmd.Dir = fmt.Sprintf("%s/%s", repoBaseDir, repo)

	stdout, err := cmd.Output()
	failOnError(fmt.Sprintf("could not create new branch: %s\n", branch), err)

	return string(stdout)
}

func setSubmoduleTo(repo string, submodule string) bool {
	filePath := fmt.Sprintf("%s/%s/.gitmodules", repoBaseDir, repo)
	input, err := ioutil.ReadFile(filePath)
	failOnError("could not open .gitmodules", err)

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "branch =") {
			lines[i] = fmt.Sprintf("\tbranch = %s", submodule)
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filePath, []byte(output), 0644)
	failOnError("could not write contents into .gitmodules", err)

	return false
}

func pushChangesToRemote(repo, branch, commit string) bool {
	var (
		cmd *exec.Cmd
		err error
	)

	if repo != "pmc-react-shared" {
		// git add
		cmd = exec.Command(commands["git"], "add", ".")
		cmd.Dir = fmt.Sprintf("%s/%s", repoBaseDir, repo)

		_, err := cmd.Output()
		cmd.Wait()
		failOnError("could stage changes, command 'git add' failed", err)

		// git commit
		cmd = exec.Command(commands["git"], "commit", "-m", commit, "--no-verify")
		cmd.Dir = fmt.Sprintf("%s/%s", repoBaseDir, repo)

		_, err = cmd.Output()
		cmd.Wait()
		failOnError("command 'git commit' failed", err)
	}

	// git push
	cmd = exec.Command(commands["git"], "push", "--set-upstream", "origin", branch, "--no-verify")
	cmd.Dir = fmt.Sprintf("%s/%s", repoBaseDir, repo)

	_, err = cmd.Output()
	cmd.Wait()
	failOnError("command 'git push' failed", err)

	return true
}
