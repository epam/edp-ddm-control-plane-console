{{- define "imageRegistry" -}}
{{- if .Values.global.imageRegistry -}}
{{- printf "%s/" .Values.global.imageRegistry -}}
{{- else -}}
{{- end -}}
{{- end }}

{{- define "edp.hostnameSuffix" -}}
{{- printf "%s-%s-%s.%s" .Release.Namespace .Values.cdPipelineName .Values.cdPipelineStageName | trunc 32 | trimSuffix "-" .Values.dnsWildcard }}
{{- end }}


{{- define "control-plane-console.hostname" -}}
{{- printf "%s-%s" .Chart.Name (include "edp.hostnameSuffix" .) }}
{{- end }}

{{- define "control-plane-console.url" -}}
{{- printf "%s%s" "https://" (include "control-plane-console.hostname" .) }}
{{- end }}

{{- define "keycloak.realm" -}}
{{- printf "%s-%s" .Release.Namespace .Values.keycloakIntegration.realm }}
{{- end -}}
