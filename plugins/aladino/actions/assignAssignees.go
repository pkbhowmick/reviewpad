// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_actions

import (
	"fmt"

	gh "github.com/reviewpad/reviewpad/v3/codehost/github"
	"github.com/reviewpad/reviewpad/v3/lang/aladino"
)

func AssignAssignees() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type: aladino.BuildFunctionType([]aladino.Type{aladino.BuildArrayOfType(aladino.BuildStringType())}, nil),
		Code: assignAssigneesCode,
	}
}

func assignAssigneesCode(e aladino.Env, args []aladino.Value) error {
	assignees := args[0].(*aladino.ArrayValue).Vals
	if len(assignees) == 0 {
		return fmt.Errorf("assignAssignees: list of assignees can't be empty")
	}

	if len(assignees) > 10 {
		return fmt.Errorf("assignAssignees: can only assign up to 10 assignees")
	}

	assigneesLogin := make([]string, len(assignees))
	for i, assignee := range assignees {
		assigneesLogin[i] = assignee.(*aladino.StringValue).Val
	}

	prNum := gh.GetPullRequestNumber(e.GetPullRequest())
	owner := gh.GetPullRequestBaseOwnerName(e.GetPullRequest())
	repo := gh.GetPullRequestBaseRepoName(e.GetPullRequest())

	_, _, err := e.GetGithubClient().AddAssignees(e.GetCtx(), owner, repo, prNum, assigneesLogin)

	return err
}
