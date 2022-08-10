// Copyright (C) 2022 Explore.dev, Unipessoal Lda - All Rights Reserved
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package github

import (
	"context"

	"github.com/google/go-github/v45/github"
)

func (c *GithubClient) GetRepositoryBranch(ctx context.Context, owner string, repo string, branch string, followRedirects bool) (*github.Branch, *github.Response, error) {
	return c.clientREST.Repositories.GetBranch(ctx, owner, repo, branch, followRedirects)
}