// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
    "net/http"
    "strings"

    "code.gitea.io/gitea/modules/gitrepo"
    api "code.gitea.io/gitea/modules/structs"
    "code.gitea.io/gitea/services/context"
)

// CompareDiff compare two branches or commits
func CompareDiff(ctx *context.APIContext) {
    // swagger:operation GET /repos/{owner}/{repo}/compare/{basehead} repository repoCompareDiff
    // ---
    // summary: Get commit comparison count
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
    //   description: compare two branches or commits
    //   type: string
    //   required: true
    // responses:
    //   "200":
    //     description: Number of commits in the comparison
    //     schema:
    //       type: object
    //       properties:
    //         total_commits:
    //           type: integer
    //           description: Number of commits in the comparison
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

    ctx.JSON(http.StatusOK, map[string]int{
        "total_commits": len(compareResult.compareInfo.Commits),
    })
}
