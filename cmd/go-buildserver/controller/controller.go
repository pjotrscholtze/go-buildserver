package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/repo"
	"github.com/pjotrscholtze/go-buildserver/models"
	"github.com/pjotrscholtze/go-buildserver/restapi/operations"
)

// ConnectControllers with the API.
func ConnectControllers(api *operations.GoBuildserverAPI, pipelineRepo repo.PipelineRepo, jobQueue repo.JobQueue) {
	api.ListPipelinesHandler = operations.ListPipelinesHandlerFunc(func(lrp operations.ListPipelinesParams) middleware.Responder {
		pipelineRepos := pipelineRepo.List()
		payload := make([]*models.Pipeline, len(pipelineRepos))
		for i, pipelineRepo := range pipelineRepos {
			var lbr repo.BuildResult
			outputLbr := make([]*models.BuildResult, 0)

			lbrs := pipelineRepo.GetLastNBuildResults(1)
			if len(lbrs) == 1 {
				lbr = lbrs[0]
				lines := make([]*models.BuildResultLine, 0)
				for _, line := range lbr.Lines {
					lines = append(lines, &models.BuildResultLine{
						Pipe: line.Pipe,
						Time: strfmt.DateTime(line.Time),
						Line: line.Line,
					})
				}
				outputLbr = append(outputLbr, &models.BuildResult{
					Reason:    lbr.Reason,
					StartTime: strfmt.DateTime(lbr.Starttime),
					Status:    string(lbr.Status),
					Lines:     lines,
				})
			}

			triggers := pipelineRepo.GetTriggers()
			payload[i] = &models.Pipeline{
				BuildScript:     pipelineRepo.GetBuildScript(),
				ForceCleanBuild: pipelineRepo.ForceCleanBuild(),
				Name:            pipelineRepo.GetName(),
				URL:             pipelineRepo.GetURL(),
				Path:            pipelineRepo.GetPath(),
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

		return operations.NewListPipelinesOK().WithPayload(payload)
	})

	api.StartPipelineHandler = operations.StartPipelineHandlerFunc(func(sbp operations.StartPipelineParams) middleware.Responder {
		// pipelineRepo.GetRepoByName(sbp.Name).Build("HTTP: " + sbp.Reason)
		jobQueue.AddQueueItem(sbp.Name, sbp.Reason, "HTTP")
		return operations.NewStartPipelineOK()
	})
	api.ListJobsHandler = operations.ListJobsHandlerFunc(func(ljp operations.ListJobsParams) middleware.Responder {
		return operations.NewListJobsOK().WithPayload(jobQueue.List())
	})

	api.GetPipelineHandler = operations.GetPipelineHandlerFunc(func(gpp operations.GetPipelineParams) middleware.Responder {
		pipeline := pipelineRepo.GetRepoBySlug(gpp.Name)
		if pipeline == nil {
			return operations.NewGetPipelineNotFound()
		}
		triggers := []*models.Trigger{}
		for _, trigger := range pipeline.GetTriggers() {
			triggers = append(triggers, &models.Trigger{
				Kind:     trigger.Kind,
				Schedule: trigger.Schedule,
			})
		}
		buildResults := jobQueue.ListAllJobsOfPipeline(gpp.Name)

		return operations.NewGetPipelineOK().WithPayload(&models.PipelineWithBuilds{
			Pipeline: &models.Pipeline{
				BuildScript:     pipeline.GetBuildScript(),
				ForceCleanBuild: pipeline.ForceCleanBuild(),
				LastBuildResult: nil,
				Name:            pipeline.GetName(),
				Path:            pipeline.GetPath(),
				Triggers:        triggers,
				URL:             pipeline.GetURL(),
			},
			Builds: buildResults,
		})
	})
}
