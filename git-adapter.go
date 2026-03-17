package main

import (
	"os/exec"
	"strconv"
	"strings"
)

type GitResult struct {
	Output string
	Err    string
	OK     bool
}

func runGit(args ...string) GitResult {
	cmd := exec.Command("git", args...)
	stdout, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return GitResult{
				Output: strings.TrimSpace(string(stdout)),
				Err:    strings.TrimSpace(string(ee.Stderr)),
				OK:     false,
			}
		}
		return GitResult{Err: err.Error(), OK: false}
	}
	return GitResult{Output: strings.TrimSpace(string(stdout)), OK: true}
}

func IsGitRepo() bool                       { return false }
func GitStatus() GitResult                  { return GitResult{} }
func GitBranch() GitResult                  { return GitResult{} }
func GitLog(n int) GitResult                { return GitResult{} }
func GitInit() GitResult                    { return GitResult{} }
func GitAdd(path string) GitResult          { return GitResult{} }
func GitCommit(message string) GitResult    { return GitResult{} }
func GitRemotes() GitResult                 { return GitResult{} }
func GitDiffStat() GitResult                { return GitResult{} }
func GitDiffStaged() GitResult              { return GitResult{} }
func HasCommits() bool                      { return false }
func IsGitIdentitySet() bool                { return false }
func GitCreateBranch(name string) GitResult { return GitResult{} }
func GitCheckout(branch string) GitResult   { return GitResult{} }
func GitMerge(branch string) GitResult      { return GitResult{} }
func GitBranchExists(name string) bool      { return false }
func GitDeleteBranch(name string) GitResult { return GitResult{} }
func GitLastMergedBranch() string           { return "" }

func itoa(n int) string { return strconv.Itoa(n) }

func sanitizeBranchName(name string) string {
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		} else if r == ' ' || r == '_' {
			result.WriteRune('-')
		}
	}
	s := result.String()
	s = strings.Trim(s, "-")
	return s
}

func resolveBaseBranch() string { return "main" }
