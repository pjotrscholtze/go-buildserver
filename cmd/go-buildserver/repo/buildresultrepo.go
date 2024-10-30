package repo

type buildResultRepo struct {
	builds []BuildResult
}

type BuildResultRepo interface {
	GetBuildResultForId(id int) *BuildResult
	ListBuildResultsForPipeline(pipeline string) map[int]BuildResult
}

func (brr *buildResultRepo) GetBuildResultForId(id int) *BuildResult {
	if id >= 0 && len(brr.builds) < id {
		return &brr.builds[id]
	}
	return nil
}
func (brr *buildResultRepo) ListBuildResultsForPipeline(pipeline string) map[int]BuildResult {
	out := map[int]BuildResult{}
	for idx, br := range brr.builds {
		if br.PipelineName != pipeline {
			continue
		}
		out[idx] = br
	}
	return out
}

func NewBuildResultRepo() BuildResultRepo {
	return &buildResultRepo{}
}
