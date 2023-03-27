package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/process"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/repo"
	"github.com/pjotrscholtze/go-buildserver/models"
	"github.com/pjotrscholtze/go-buildserver/restapi/operations"
)

func fmtPipeName(pt process.PipeType) string {
	if pt == process.STDOUT {
		return "STDOUT"
	}
	return "STDERR"
}

// ConnectControllers with the API.
func ConnectControllers(api *operations.GoBuildserverAPI, buildRepo repo.BuildRepo) {
	api.ListReposHandler = operations.ListReposHandlerFunc(func(lrp operations.ListReposParams) middleware.Responder {
		buildRepos := buildRepo.List()
		payload := make([]*models.Repo, len(buildRepos))
		for i, buildRepo := range buildRepos {
			var lbr repo.BuildResult
			outputLbr := make([]*models.BuildResult, 0)

			lbrs := buildRepo.GetLastNBuildResults(1)
			if len(lbrs) == 1 {
				lbr = lbrs[0]
				lines := make([]*models.BuildResultLine, 0)
				for _, line := range lbr.Lines() {
					lines = append(lines, &models.BuildResultLine{
						Pipe: fmtPipeName(line.Pipe()),
						Time: strfmt.DateTime(line.Time()),
						Line: line.Line(),
					})

				}
				outputLbr = append(outputLbr, &models.BuildResult{
					Reason:    lbr.Reason(),
					StartTime: strfmt.DateTime(lbr.Starttime()),
					Status:    string(lbr.Status()),
					Lines:     lines,
				})
			}

			triggers := buildRepo.GetTriggers()
			payload[i] = &models.Repo{
				BuildScript:     buildRepo.GetBuildScript(),
				ForceCleanBuild: buildRepo.ForceCleanBuild(),
				Name:            buildRepo.GetName(),
				URL:             buildRepo.GetURL(),
				LastBuildResult: outputLbr,
				Triggers:        make([]*models.Trigger, len(triggers)),
			}
			for y, trigger := range triggers {
				payload[i].Triggers[y] = &models.Trigger{
					Kind:     trigger.Kind,
					Schedule: trigger.Schedule,
				}
			}
		}
		return operations.NewListReposOK().WithPayload(payload)
	})
	api.StartBuildHandler = operations.StartBuildHandlerFunc(func(sbp operations.StartBuildParams) middleware.Responder {
		buildRepo.GetRepoByName(sbp.Name).Build("HTTP: " + sbp.Reason)
		return operations.NewStartBuildOK()
	})
}
