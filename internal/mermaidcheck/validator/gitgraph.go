package validator

import (
	"github.com/riclib/mmdr-go/internal/mermaidcheck/ast"
)

// GitGraphRule is a validation rule for git graph diagrams.
type GitGraphRule interface {
	Validate(diagram *ast.GitGraphDiagram) []*ValidationError
}

// ValidateGitGraph runs validation rules on a git graph diagram.
func ValidateGitGraph(diagram *ast.GitGraphDiagram, strict bool) []*ValidationError {
	rules := GitGraphDefaultRules()
	if strict {
		rules = GitGraphStrictRules()
	}

	var errors []*ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(diagram)...)
	}
	return errors
}

// GitGraphDefaultRules returns the default validation rules for git graph diagrams.
func GitGraphDefaultRules() []GitGraphRule {
	return []GitGraphRule{
		&NoDuplicateBranchNamesRule{},
		&ValidBranchReferencesRule{},
		&ValidCommitReferencesRule{},
		&ValidCommitTypeRule{},
	}
}

// GitGraphStrictRules returns strict validation rules for git graph diagrams.
func GitGraphStrictRules() []GitGraphRule {
	rules := GitGraphDefaultRules()
	// Add strict-only rules here if needed
	return rules
}

// NoDuplicateBranchNamesRule checks for duplicate branch names.
type NoDuplicateBranchNamesRule struct{}

// Validate checks that all branch names are unique.
func (r *NoDuplicateBranchNamesRule) Validate(diagram *ast.GitGraphDiagram) []*ValidationError {
	checker := NewDuplicateChecker("branch")
	var errors []*ValidationError

	for _, op := range diagram.Operations {
		if op.Type == "branch" {
			if err := checker.Check(op.BranchName, op.Pos); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

// ValidBranchReferencesRule checks that checkout and merge operations reference existing branches.
type ValidBranchReferencesRule struct{}

// Validate checks that all branch references are valid.
func (r *ValidBranchReferencesRule) Validate(diagram *ast.GitGraphDiagram) []*ValidationError {
	branchChecker := NewReferenceChecker("branch")
	var errors []*ValidationError

	// Add main branch (always exists)
	if diagram.MainBranchName != "" {
		branchChecker.Add(diagram.MainBranchName)
	} else {
		branchChecker.Add("main") // default main branch
	}

	// First pass: collect all branch definitions
	for _, op := range diagram.Operations {
		if op.Type == "branch" {
			branchChecker.Add(op.BranchName)
		}
	}

	// Second pass: validate references
	for _, op := range diagram.Operations {
		switch op.Type {
		case "checkout":
			if err := branchChecker.Check(op.BranchName, op.Pos, "checkout"); err != nil {
				errors = append(errors, err)
			}
		case "merge":
			if err := branchChecker.Check(op.BranchName, op.Pos, "merge"); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

// ValidCommitReferencesRule checks that cherry-pick operations reference existing commits.
type ValidCommitReferencesRule struct{}

// Validate checks that all commit references are valid.
func (r *ValidCommitReferencesRule) Validate(diagram *ast.GitGraphDiagram) []*ValidationError {
	commitChecker := NewReferenceChecker("commit")
	var errors []*ValidationError

	// First pass: collect all commit IDs
	for _, op := range diagram.Operations {
		if (op.Type == "commit" || op.Type == "merge") && op.ID != "" {
			commitChecker.Add(op.ID)
		}
	}

	// Second pass: validate cherry-pick references
	for _, op := range diagram.Operations {
		if op.Type == "cherry-pick" {
			if err := commitChecker.Check(op.ParentID, op.Pos, "cherry-pick"); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

// ValidCommitTypeRule checks that commit types are valid.
type ValidCommitTypeRule struct{}

// Validate checks that all commit types are NORMAL, REVERSE, or HIGHLIGHT.
func (r *ValidCommitTypeRule) Validate(diagram *ast.GitGraphDiagram) []*ValidationError {
	validator := NewEnumValidator("commit type", "NORMAL", "REVERSE", "HIGHLIGHT")
	var errors []*ValidationError

	for _, op := range diagram.Operations {
		if (op.Type == "commit" || op.Type == "merge") && op.CommitType != "" {
			if err := validator.Check(op.CommitType, op.Pos); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}
