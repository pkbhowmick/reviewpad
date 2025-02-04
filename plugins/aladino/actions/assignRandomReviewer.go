// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_actions

import (
	"fmt"

	"github.com/google/go-github/v45/github"
	gh "github.com/reviewpad/reviewpad/v3/codehost/github"
	"github.com/reviewpad/reviewpad/v3/lang/aladino"
	"github.com/reviewpad/reviewpad/v3/utils"
)

func AssignRandomReviewer() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type: aladino.BuildFunctionType([]aladino.Type{}, nil),
		Code: assignRandomReviewerCode,
	}
}

func assignRandomReviewerCode(e aladino.Env, _ []aladino.Value) error {
	prNum := gh.GetPullRequestNumber(e.GetPullRequest())
	owner := gh.GetPullRequestBaseOwnerName(e.GetPullRequest())
	repo := gh.GetPullRequestBaseRepoName(e.GetPullRequest())

	ghPrRequestedReviewers, err := e.GetGithubClient().GetPullRequestReviewers(e.GetCtx(), owner, repo, prNum, &github.ListOptions{})
	if err != nil {
		return err
	}

	// When there's already assigned reviewers, do nothing
	totalRequestReviewers := len(ghPrRequestedReviewers.Users)
	if totalRequestReviewers > 0 {
		return nil
	}

	ghUsers, err := e.GetGithubClient().GetIssuesAvailableAssignees(e.GetCtx(), owner, repo)
	if err != nil {
		return err
	}

	filteredGhUsers := []*github.User{}

	for i := range ghUsers {
		if ghUsers[i].GetLogin() != e.GetPullRequest().GetUser().GetLogin() {
			filteredGhUsers = append(filteredGhUsers, ghUsers[i])
		}
	}

	if len(filteredGhUsers) == 0 {
		return fmt.Errorf("can't assign a random user because there is no users")
	}

	lucky := utils.GenerateRandom(len(filteredGhUsers))
	ghUser := filteredGhUsers[lucky]

	_, _, err = e.GetGithubClient().RequestReviewers(e.GetCtx(), owner, repo, prNum, github.ReviewersRequest{
		Reviewers: []string{ghUser.GetLogin()},
	})

	return err
}
