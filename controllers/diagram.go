/*
 * Copyright 2020 EPAM Systems.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"ddm-admin-console/console"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	pipelineService "ddm-admin-console/service/cd_pipeline"
	"ddm-admin-console/util"
	"encoding/json"
	"github.com/astaxie/beego"
	"go.uber.org/zap"
)

type DiagramController struct {
	beego.Controller
	CodebaseService service.CodebaseService
	PipelineService pipelineService.CDPipelineService
}

const diagramPageType = "diagram"

func (c *DiagramController) GetDiagramPage() {
	log.Debug("start rendering delivery_dashboard_diagram.html page")
	cJSON, err := c.getCodebasesJSON()
	if err != nil {
		log.Error("couldn't get codebases from db", zap.Error(err))
		c.Abort("500")
		return
	}

	pJSON, err := c.getPipelinesJSON()
	if err != nil {
		log.Error("couldn't get pipelines from db", zap.Error(err))
		c.Abort("500")
		return
	}

	sJSON, err := c.getCodebaseDokcerStreamsJSON()
	if err != nil {
		log.Error("couldn't get codebase docker streams from db", zap.Error(err))
		c.Abort("500")
		return
	}

	c.Data["Username"] = c.Ctx.Input.Session("username")
	c.Data["EDPVersion"] = console.EDPVersion
	c.Data["CodebasesJson"] = cJSON
	c.Data["PipelinesJson"] = pJSON
	c.Data["CodebaseDockerStreamsJson"] = sJSON
	c.Data["DiagramPageEnabled"] = console.DiagramPageEnabled
	c.Data["Type"] = diagramPageType
	c.Data["BasePath"] = console.BasePath
	c.TplName = "delivery_dashboard_diagram.html"
}

func (c *DiagramController) getCodebasesJSON() (*string, error) {
	codebases, err := c.CodebaseService.GetCodebasesByCriteria(query.CodebaseCriteria{})
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(codebases)
	if err != nil {
		return nil, err
	}
	return util.GetStringP(string(buf)), nil
}

func (c *DiagramController) getPipelinesJSON() (*string, error) {
	pipelines, err := c.PipelineService.GetAllPipelines(query.CDPipelineCriteria{})
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(pipelines)
	if err != nil {
		return nil, err
	}
	return util.GetStringP(string(buf)), nil
}

func (c *DiagramController) getCodebaseDokcerStreamsJSON() (*string, error) {
	streams, err := c.PipelineService.GetAllCodebaseDockerStreams()
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(streams)
	if err != nil {
		return nil, err
	}
	return util.GetStringP(string(buf)), nil
}
