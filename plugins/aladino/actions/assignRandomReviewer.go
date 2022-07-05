// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_actions

import (
	"fmt"

	"github.com/google/go-github/v42/github"
	"github.com/reviewpad/reviewpad/v2/lang/aladino"
	"github.com/reviewpad/reviewpad/v2/utils"
)

func AssignRandomReviewer() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type: aladino.BuildFunctionType([]aladino.Type{}, nil),
		Code: assignRandomReviewerCode,
	}
}

func assignRandomReviewerCode(e aladino.Env, _ []aladino.Value) error {
	prNum := utils.GetPullRequestNumber(e.GetPullRequest())
	owner := utils.GetPullRequestOwnerName(e.GetPullRequest())
	repo := utils.GetPullRequestRepoName(e.GetPullRequest())

	ghPr, _, err := e.GetClient().PullRequests.Get(e.GetCtx(), owner, repo, prNum)
	if err != nil {
		return err
	}

	// When there's already assigned reviewers, do nothing
	totalRequestReviewers := len(ghPr.RequestedReviewers)
	if totalRequestReviewers > 0 {
		return nil
	}

	ghUsers, _, err := e.GetClient().Repositories.ListCollaborators(e.GetCtx(), owner, repo, nil)
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

	_, _, err = e.GetClient().PullRequests.RequestReviewers(e.GetCtx(), owner, repo, prNum, github.ReviewersRequest{
		Reviewers: []string{ghUser.GetLogin()},
	})

	return err
}