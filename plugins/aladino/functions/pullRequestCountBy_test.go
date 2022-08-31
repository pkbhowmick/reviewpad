// Copyright (C) 2022 Explore.dev, Unipessoal Lda - All Rights Reserved
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_functions_test

import (
	"net/http"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/reviewpad/reviewpad/v3/lang/aladino"
	plugins_aladino "github.com/reviewpad/reviewpad/v3/plugins/aladino"
	"github.com/stretchr/testify/assert"
)

var pullRequestCountBy = plugins_aladino.PluginBuiltIns().Functions["pullRequestCountBy"].Code

func TestPullRequestCountBy_WhenListIssuesByRepoFails(t *testing.T) {
	failMessage := "ListListIssuesByRepoRequestFail"
	mockedEnv := aladino.MockDefaultEnv(
		t,
		[]mock.MockBackendOption{
			mock.WithRequestMatchHandler(
				mock.GetReposIssuesByOwnerByRepo,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mock.WriteError(
						w,
						http.StatusInternalServerError,
						failMessage,
					)
				}),
			),
		},
		nil,
		aladino.MockBuiltIns(),
		nil,
	)

	args := []aladino.Value{aladino.BuildStringValue(""), aladino.BuildStringValue("")}
	gotTotal, err := pullRequestCountBy(mockedEnv, args)

	assert.Nil(t, gotTotal)
	assert.Equal(t, err.(*github.ErrorResponse).Message, failMessage)
}

func TestPullRequestCountBy(t *testing.T) {
	firstIssue := &github.Issue{
		Title: github.String("First Issue"),
		State: github.String("open"),
	}

	secondIssue := &github.Issue{
		Title: github.String("Second Issue"),
		PullRequestLinks: &github.PullRequestLinks{
			URL: github.String("pull-request-link"),
		},
		State: github.String("closed"),
	}

	thirdIssue := &github.Issue{
		Title: github.String("Third Issue"),
		State: github.String("closed"),
		User: &github.User{
			Login: github.String("steve"),
		},
	}

	tests := map[string]struct {
		args    []aladino.Value
		issues  []*github.Issue
		wantVal aladino.Value
	}{
		"default values": {
			args:    []aladino.Value{aladino.BuildStringValue(""), aladino.BuildStringValue("")},
			issues:  []*github.Issue{firstIssue, secondIssue, thirdIssue},
			wantVal: aladino.BuildIntValue(1),
		},
		"only closed pull requests": {
			args:    []aladino.Value{aladino.BuildStringValue(""), aladino.BuildStringValue("closed")},
			issues:  []*github.Issue{firstIssue, secondIssue},
			wantVal: aladino.BuildIntValue(1),
		},
		"only pull requests by steve": {
			args:    []aladino.Value{aladino.BuildStringValue("steve"), aladino.BuildStringValue("")},
			issues:  []*github.Issue{firstIssue, thirdIssue},
			wantVal: aladino.BuildIntValue(0),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockedEnv := aladino.MockDefaultEnv(
				t,
				[]mock.MockBackendOption{
					mock.WithRequestMatchHandler(
						mock.GetReposIssuesByOwnerByRepo,
						http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
							w.Write(mock.MustMarshal(test.issues))
						}),
					),
				},
				nil,
				aladino.MockBuiltIns(),
				nil,
			)

			gotVal, err := pullRequestCountBy(mockedEnv, test.args)

			assert.Nil(t, err)
			assert.Equal(t, test.wantVal, gotVal)
		})
	}
}