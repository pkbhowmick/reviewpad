// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino

import (
	"fmt"
	"log"

	"github.com/google/go-github/v42/github"
	"github.com/reviewpad/reviewpad/lang/aladino"
	"github.com/reviewpad/reviewpad/utils"
)

/*
reviewpad-an: builtin-docs

## addLabel

**Description**:

Adds a label to the pull request.

This built-in assumes that the label has been created. Otherwise, it returns an error.

**Parameters**:

| variable | type   | description       |
| -------- | ------ | ----------------- |
| `name`   | string | name of the label |

**Return value**:

Error if the label does not exist in the repository.

**Examples**:

```yml
$addLabel("bug")
```

A `revy.yml` example:

```yml
protectionGates:
  - name: label-small-pull-request
    description: Label small pull request
    patchRules:
      - rule: isSmall
    actions:
      - $addLabel("small")
```
*/
func addLabel() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type: aladino.BuildFunctionType([]aladino.Type{aladino.BuildStringType()}, nil),
		Code: addLabelCode,
	}
}

func addLabelCode(e *aladino.EvalEnv, args []aladino.Value) error {
	if len(args) != 1 {
		return fmt.Errorf("addLabel: expecting 1 argument, got %v", len(args))
	}

	labelVal := args[0]
	if !labelVal.HasKindOf(aladino.STRING_VALUE) {
		return fmt.Errorf("addLabel: expecting string argument, got %v", labelVal.Kind())
	}

	label := labelVal.(*aladino.StringValue).Val

	prNum := utils.GetPullRequestNumber(e.PullRequest)
	owner := utils.GetPullRequestOwnerName(e.PullRequest)
	repo := utils.GetPullRequestRepoName(e.PullRequest)

	_, _, err := e.Client.Issues.GetLabel(e.Ctx, owner, repo, label)
	if err != nil {
		return err
	}

	_, _, err = e.Client.Issues.AddLabelsToIssue(e.Ctx, owner, repo, prNum, []string{label})

	return err
}

/*
reviewpad-an: builtin-docs

## assignRandomReviewer

**Description**:

Assigns a random user of the GitHub organization as the reviewer.
This action will always pick a user different than the author of the pull request.

However, if the pull request already has a reviewer, nothing happens. This is to prevent
adding a reviewer each time the pull request is updated.

When there's no reviewers to assign to, an error is returned.

**Parameters**:

*none*

**Return value**:

*none*

**Examples**:

```yml
$assignRandomReviewer()
```

A `revy.yml` example:

```yml
protectionGates:
  - name: assign-random-reviewer
    description: Assign random reviewer
    patchRules:
      - rule: tautology
    actions:
      - $assignRandomReviewer()
```
*/
func assignRandomReviewer() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type: aladino.BuildFunctionType([]aladino.Type{}, nil),
		Code: assignRandomReviewerCode,
	}
}

func assignRandomReviewerCode(e *aladino.EvalEnv, _ []aladino.Value) error {
	prNum := utils.GetPullRequestNumber(e.PullRequest)
	owner := utils.GetPullRequestOwnerName(e.PullRequest)
	repo := utils.GetPullRequestRepoName(e.PullRequest)

	ghPr, _, err := e.Client.PullRequests.Get(e.Ctx, owner, repo, prNum)
	if err != nil {
		return err
	}

	// When there's already assigned reviewers, do nothing
	totalRequestReviewers := len(ghPr.RequestedReviewers)
	if totalRequestReviewers > 0 {
		return nil
	}

	ghUsers, _, err := e.Client.Repositories.ListCollaborators(e.Ctx, owner, repo, nil)
	if err != nil {
		return err
	}

	filteredGhUsers := []*github.User{}

	for i := range ghUsers {
		if ghUsers[i].GetLogin() != ghPr.GetUser().GetLogin() {
			filteredGhUsers = append(filteredGhUsers, ghUsers[i])
		}
	}

	if len(filteredGhUsers) == 0 {
		return fmt.Errorf("can't assign a random user because there is no users")
	}

	lucky := utils.GenerateRandom(len(filteredGhUsers))
	ghUser := filteredGhUsers[lucky]

	_, _, err = e.Client.PullRequests.RequestReviewers(e.Ctx, owner, repo, prNum, github.ReviewersRequest{
		Reviewers: []string{ghUser.GetLogin()},
	})

	return err
}

/*
reviewpad-an: builtin-docs

## assignReviewer

**Description**:

Assigns a defined amount of reviewers to the pull request from the provided list of reviewers.

When there's not enough reviewers to assign to, an warning is returned.

If a reviewer from the defined list has performed a review, his review will re-requested.

**Parameters**:

| variable           | type     | description                                                       |
| ------------------ | -------- | ----------------------------------------------------------------- |
| `reviewers`        | []string | list of GitHub logins to select from                              |
| `total` (optional) | int      | total of reviewers to assign. by default assigns to all reviewers |

**Return value**:

*none*

**Examples**:

```yml
$assignReviewer(["john", "marie", "peter"], 2)
```

A `revy.yml` example:

```yml
protectionGates:
  - name: review-code-from-new-joiners
    description: Assign senior reviewers to PRs from new joiners
    patchRules:
      - rule: authoredByJunior
    actions:
      - $assignReviewer($group("seniors"), 2)
```
*/
func assignReviewer() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type: aladino.BuildFunctionType([]aladino.Type{aladino.BuildArrayOfType(aladino.BuildStringType()), aladino.BuildIntType()}, nil),
		Code: assignReviewerCode,
	}
}

func assignReviewerCode(e *aladino.EvalEnv, args []aladino.Value) error {
	if len(args) < 1 {
		return fmt.Errorf("assignReviewer: expecting at least 1 argument")
	}

	arg := args[0]
	if !arg.HasKindOf(aladino.ARRAY_VALUE) {
		return fmt.Errorf("assignReviewer: requires array argument, got %v", arg.Kind())
	}

	if !args[1].HasKindOf(aladino.INT_VALUE) {
		return fmt.Errorf("assignReviewer: the parameter total is required to be an int, instead got %v", args[1].Kind())
	}

	totalRequiredReviewers := args[1].(*aladino.IntValue).Val

	availableReviewers := arg.(*aladino.ArrayValue).Vals

	for _, reviewer := range availableReviewers {
		if !reviewer.HasKindOf(aladino.STRING_VALUE) {
			return fmt.Errorf("assignReviewer: requires array of strings, got array with value of %v", reviewer.Kind())
		}
	}

	// Remove pull request author from provided reviewers list
	for index, reviewer := range availableReviewers {
		if reviewer.(*aladino.StringValue).Val == *e.PullRequest.User.Login {
			availableReviewers = append(availableReviewers[:index], availableReviewers[index+1:]...)
			break
		}
	}

	totalAvailableReviewers := len(availableReviewers)
	if totalRequiredReviewers > totalAvailableReviewers {
		log.Printf("assignReviewer: total required reviewers %v exceeds the total available reviewers %v", totalRequiredReviewers, totalAvailableReviewers)
		totalRequiredReviewers = totalAvailableReviewers
	}

	prNum := utils.GetPullRequestNumber(e.PullRequest)
	owner := utils.GetPullRequestOwnerName(e.PullRequest)
	repo := utils.GetPullRequestRepoName(e.PullRequest)

	reviewers := []string{}

	reviews, _, err := e.Client.PullRequests.ListReviews(e.Ctx, owner, repo, prNum, nil)
	if err != nil {
		return err
	}

	// Re-request current reviewers if mention on the provided reviewers list
	for _, review := range reviews {
		for index, availableReviewer := range availableReviewers {
			if availableReviewer.(*aladino.StringValue).Val == *review.User.Login {
				totalRequiredReviewers--
				reviewers = append(reviewers, *review.User.Login)
				availableReviewers = append(availableReviewers[:index], availableReviewers[index+1:]...)
				break
			}
		}
	}

	// Skip current requested reviewers if mention on the provided reviewers list
	currentRequestedReviewers := e.PullRequest.RequestedReviewers
	for _, requestedReviewer := range currentRequestedReviewers {
		for index, availableReviewer := range availableReviewers {
			if availableReviewer.(*aladino.StringValue).Val == *requestedReviewer.Login {
				totalRequiredReviewers--
				availableReviewers = append(availableReviewers[:index], availableReviewers[index+1:]...)
				break
			}
		}
	}

	// Select random reviewers from the list of all provided reviewers
	for i := 0; i < totalRequiredReviewers; i++ {
		selectedElementIndex := utils.GenerateRandom(len(availableReviewers))

		selectedReviewer := availableReviewers[selectedElementIndex]
		availableReviewers = append(availableReviewers[:selectedElementIndex], availableReviewers[selectedElementIndex+1:]...)

		reviewers = append(reviewers, selectedReviewer.(*aladino.StringValue).Val)
	}

	if len(reviewers) == 0 {
		log.Printf("assignReviewer: skipping request reviewers. the pull request already has reviewers")
		return nil
	}

	_, _, err = e.Client.PullRequests.RequestReviewers(e.Ctx, owner, repo, prNum, github.ReviewersRequest{
		Reviewers: reviewers,
	})

	return err
}

/*
reviewpad-an: builtin-docs

## merge

**Description**:

Merge a pull request with a specific merge method.

By default, if no parameter is provided, it will perform a standard git merge.

| :warning: Requires a GitHub token :warning: |
|---------------------------------------------|

By default a GitHub action does not have permission to access organization members.

Because of that, in order for the function `team` to work we need to provide a GitHub token to the Reviewpad action.

[Please follow this link to know more](https://docs.reviewpad.com/docs/install-github-action-with-github-token).

**Parameters**:

| variable       | type   | description                            |
| -------------- | ------ | -------------------------------------- |
| `method`       | string | merge method (merge, rebase or squash) |

**Return value**:

*none*

**Examples**:

```yml
$merge()
```

A `revy.yml` example:

```yml
protectionGates:
  - name: auto-merge-small-pull-request
    description: Auto-merge small pull request
    patchRules:
      - rule: isSmall
    actions:
      - $merge()
```
*/
func merge() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type: aladino.BuildFunctionType([]aladino.Type{aladino.BuildStringType()}, nil),
		Code: mergeCode,
	}
}

func mergeCode(e *aladino.EvalEnv, args []aladino.Value) error {
	prNum := utils.GetPullRequestNumber(e.PullRequest)
	owner := utils.GetPullRequestOwnerName(e.PullRequest)
	repo := utils.GetPullRequestRepoName(e.PullRequest)

	mergeMethod, err := parseMergeMethod(args)
	if err != nil {
		return err
	}

	_, _, err = e.Client.PullRequests.Merge(e.Ctx, owner, repo, prNum, "Merged by Reviewpad", &github.PullRequestOptions{
		MergeMethod: mergeMethod,
	})
	return err
}

func parseMergeMethod(args []aladino.Value) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("merge: received two arguments")
	}

	if len(args) == 0 {
		return "merge", nil
	}

	arg := args[0]
	if arg.HasKindOf(aladino.STRING_VALUE) {
		mergeMethod := arg.(*aladino.StringValue).Val
		switch mergeMethod {
		case "merge", "rebase", "squash":
			return mergeMethod, nil
		default:
			return "", fmt.Errorf("merge: unexpected argument %v", mergeMethod)
		}
	} else {
		return "", fmt.Errorf("merge: expects string argument")
	}
}

/*
reviewpad-an: builtin-docs

## removeLabel

**Description**:

Removes a label applied to a pull request.

If the label is not applied to the pull request then nothing happens.

This built-in assumes that the label has been created. Otherwise, it returns an error.

**Parameters**:

| variable | type   | description       |
| -------- | ------ | ----------------- |
| `name`   | string | name of the label |

**Return value**:

Error if the label does not exist in the repository.

**Examples**:

```yml
$removeLabel("bug")
```

A `revy.yml` example:

```yml
protectionGates:
  - name: remove-small-label-in-pull-request
    description: Remove small label applied to pull request
    patchRules:
      - rule: isNotSmall
    actions:
      - $removeLabel("small")
```
*/
func removeLabel() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type: aladino.BuildFunctionType([]aladino.Type{aladino.BuildStringType()}, nil),
		Code: removeLabelCode,
	}
}

func removeLabelCode(e *aladino.EvalEnv, args []aladino.Value) error {
	if len(args) != 1 {
		return fmt.Errorf("removeLabel: expecting 1 argument, got %v", len(args))
	}

	labelVal := args[0]
	if !labelVal.HasKindOf(aladino.STRING_VALUE) {
		return fmt.Errorf("removeLabel: expecting string argument, got %v", labelVal.Kind())
	}

	label := labelVal.(*aladino.StringValue).Val

	prNum := utils.GetPullRequestNumber(e.PullRequest)
	owner := utils.GetPullRequestOwnerName(e.PullRequest)
	repo := utils.GetPullRequestRepoName(e.PullRequest)

	_, _, err := e.Client.Issues.GetLabel(e.Ctx, owner, repo, label)
	if err != nil {
		return err
	}

	var labelIsAppliedToPullRequest bool = false
	for _, ghLabel := range e.PullRequest.Labels {
		if ghLabel.GetName() == label {
			labelIsAppliedToPullRequest = true
			break
		}
	}

	if !labelIsAppliedToPullRequest {
		return nil
	}

	_, err = e.Client.Issues.RemoveLabelForIssue(e.Ctx, owner, repo, prNum, label)

	return err
}