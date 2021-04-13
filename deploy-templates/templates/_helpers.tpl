{{/*
Expand the name of the chart.
*/}}
{{- define "jenkins-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "jenkins-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "jenkins-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "dockerProxyRegistry" -}}
{{- if .Values.dockerProxyRegistry -}}
{{- printf "%s/" .Values.dockerProxyRegistry -}}
{{- else -}}
{{- end -}}
{{- end }}

{{- define "edp.hostnameSuffix" -}}
{{- printf "%s-%s.%s" .Values.cdPipelineName .Values.cdPipelineStageName .Values.dnsWildcard }}
{{- end }}

{{- define "keycloak.realm" -}}
{{- printf "%s-%s" .Values.namespace .Values.keycloak.realm }}
{{- end -}}

{{- define "jenkinsDeployer.hostname" -}}
{{- $hostname := default (include "jenkins-operator.fullname" .) }}
{{- printf "%s-%s" $hostname (include "edp.hostnameSuffix" .) }}
{{- end }}

{{- define "jenkinsDeployer.url" -}}
{{- printf "%s%s" "https://" (include "jenkinsDeployer.hostname" .) }}
{{- end }}

{{- define "jenkins.edpSharedLibraries" -}}
{{- $defaultBaseUrl := printf "https://gerrit-%s.%s" .Values.edpProject .Values.dnsWildcard -}}
{{- $baseUrl := .Values.jenkins.edpSharedLibraries.baseUrl | default $defaultBaseUrl  -}}
- id: Stages
  name: edp-library-stages
  repository: "{{ $baseUrl }}/{{ .Values.jenkins.edpSharedLibraries.stages }}.git"
  version: {{ .Values.jenkins.edpSharedLibraries.version }}
- id: Pipelines
  name: edp-library-pipelines
  repository: "{{ $baseUrl }}/{{ .Values.jenkins.edpSharedLibraries.pipelines }}.git"
  version: {{ .Values.jenkins.edpSharedLibraries.version }}
{{- end }}
