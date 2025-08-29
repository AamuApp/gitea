// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
    "net/http"
    "strings"

    user_model "code.gitea.io/gitea/models/user"
    "code.gitea.io/gitea/modules/gitrepo"
    api "code.gitea.io/gitea/modules/structs"
    "code.gitea.io/gitea/services/context"
    "code.gitea.io/gitea/services/convert"
)

// CompareDiff compare two branches or commits
func CompareDiff(ctx *context.APIContext) {
    // swagger:operation GET /repos/{owner}/{repo}/compare/{basehead} repository repoCompareDiff
    // ---
    // summary: Get commit comparison information
    // produces:
    // - application/json
    // parameters:
    // - name: owner
    //   in: path
    //   description: owner of the repo
    //   type: string
    //   required: true
    // - name: repo
    //   in: path
    //   description: name of the repo
    //   type: string
    //   required: true
    // - name: basehead
    //   in: path
    //   description: compare two branches or commits (e.g., base...head)
    //   type: string
    //   required: true
    // - name: include_commits
    //   in: query
    //   description: whether to include the list of commits in the response
    //   type: boolean
    //   required: false
    // responses:
    //   "200":
    //     description: Commit comparison information
    //     schema:
    //       type: object
    //       properties:
    //         total_commits:
    //           type: integer
    //           description: Number of commits in the comparison
    //         commits:
    //           type: array
    //           items:
    //             "$ref": "#/definitions/Commit"
    //           description: List of commits (included only if include_commits=true)
    //   "404":
    //     "$ref": "#/responses/notFound"

    if ctx.Repo.GitRepo == nil {
        var err error
        ctx.Repo.GitRepo, err = gitrepo.RepositoryFromRequestContextOrOpen(ctx, ctx.Repo.Repository)
        if err != nil {
            ctx.APIErrorInternal(err)
            return
        }
    }

    infoPath := ctx.PathParam("*")
    infos := []string{ctx.Repo.Repository.DefaultBranch, ctx.Repo.Repository.DefaultBranch}
    if infoPath != "" {
        infos = strings.SplitN(infoPath, "...", 2)
        if len(infos) != 2 {
            if infos = strings.SplitN(infoPath, "..", 2); len(infos) != 2 {
                infos = []string{ctx.Repo.Repository.DefaultBranch, infoPath}
            }
        }
    }

    compareResult, closer := parseCompareInfo(ctx, api.CreatePullRequestOption{Base: infos[0], Head: infos[1]})
    if ctx.Written() {
        return
    }
    defer closer()

    // Check include_commits query parameter
    includeCommits := ctx.FormBool("include_commits")

    if includeCommits {
        // Original behavior: include total_commits and commits
        verification := ctx.FormString("verification") == "" || ctx.FormBool("verification")
        files := ctx.FormString("files") == "" || ctx.FormBool("files")

        apiCommits := make([]*api.Commit, 0, len(compareResult.compareInfo.Commits))
        userCache := make(map[string]*user_model.User)
        for i := 0; i < len(compareResult.compareInfo.Commits); i++ {
            apiCommit, err := convert.ToCommit(ctx, ctx.Repo.Repository, ctx.Repo.GitRepo, compareResult.compareInfo.Commits[i], userCache,
                convert.ToCommitOptions{
                    Stat:         true,
                    Verification: verification,
                    Files:        files,
                })
            if err != nil {
                ctx.APIErrorInternal(err)
                return
            }
            apiCommits = append(apiCommits, apiCommit)
        }

        ctx.JSON(http.StatusOK, &api.Compare{
            TotalCommits: len(compareResult.compareInfo.Commits),
            Commits:      apiCommits,
        })
    } else {
        // New behavior: return only total_commits
        ctx.JSON(http.StatusOK, map[string]int{
            "total_commits": len(compareResult.compareInfo.Commits),
        })
    }
}