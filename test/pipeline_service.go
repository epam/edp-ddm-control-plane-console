package test

import "ddm-admin-console/models/query"

type MockPipelineService struct {
	GetAllPipelinesResult []*query.CDPipeline
	GetAllPipelinesError  error

	GetAllCodebaseDockerStreamsResult []string
	GetAllCodebaseDockerStreamsError  error
}

func (m MockPipelineService) GetAllPipelines(criteria query.CDPipelineCriteria) ([]*query.CDPipeline, error) {
	return m.GetAllPipelinesResult, m.GetAllPipelinesError
}

func (m MockPipelineService) GetAllCodebaseDockerStreams() ([]string, error) {
	return m.GetAllCodebaseDockerStreamsResult, m.GetAllCodebaseDockerStreamsError
}
