package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

var (
	selectedBranches []string
	confirmed        bool
)

func main() {
	branches, err := getBranches()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	accessible, _ := strconv.ParseBool(os.Getenv("ACCESSIBLE"))
	var options []huh.Option[string]
	for _, branch := range branches {
		options = append(options, huh.NewOption(branch, branch))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().Options(options...).Title("What branches do you want to delete?").Value(&selectedBranches),
		),

		huh.NewGroup(huh.NewConfirm().Title("Are you sure you want to delete the selected branches?").Affirmative("Yes").Negative("No").Value(&confirmed)),
	).WithAccessible(accessible)

	err = form.Run()
	if err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}

	if confirmed {
		runDeletion(accessible)
	}
}

func runDeletion(accessible bool) {
	branches, err := deleteBranches(selectedBranches)
	if err != nil {
		newForm := huh.NewForm(
			huh.NewGroup(huh.NewConfirm().Title("It appears that normal deletion didn't work. Do you want to force delete the branches?").Affirmative("Yes").Negative("No").Value(&confirmed)),
		).WithAccessible(accessible)
		err = newForm.Run()
		if err != nil {
			fmt.Println("Uh oh:", err)
			os.Exit(1)
		}

		if confirmed {
			err = forceDeleteBranches(branches)
			if err != nil {
				fmt.Println("Uh oh:", err)
				os.Exit(1)
			}
		}
	}
}

func getBranches() ([]string, error) {
	cmd := exec.Command("git", "branch")

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return []string{}, err
	}
	output := out.String()
	result := strings.Split(output, "\n")
	var branches []string

	for _, item := range result {
		if item == "" || strings.HasPrefix(item, "*") {
			continue
		}
		branches = append(branches, item)
	}

	if len(branches) == 0 {
		return []string{}, fmt.Errorf("there are no branches I could delete here")
	}

	return branches, nil
}

func deleteBranches(branches []string) ([]string, error) {
	for idx, branch := range branches {
		cmd := exec.Command("git", "branch", "-d", branch)

		if errors.Is(cmd.Err, exec.ErrDot) {
			cmd.Err = nil
		}

		err := cmd.Run()
		if err != nil {
			return branches[idx:], err
		}
	}

	return []string{}, nil
}

func forceDeleteBranches(branches []string) error {
	for _, branch := range branches {
		cmd := exec.Command("git", "branch", "-D", branch)

		if errors.Is(cmd.Err, exec.ErrDot) {
			cmd.Err = nil
		}

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
