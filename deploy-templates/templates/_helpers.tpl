{{- define "imageRegistry" -}}
{{- if .Values.global.imageRegistry -}}
{{- printf "%s/" .Values.global.imageRegistry -}}
{{- else -}}
{{- end -}}
{{- end }}

{{- define "control-plane-console.hostname" -}}
{{- $hostname := printf "%s-%s" "control-plane-console" .Release.Namespace }}
{{- printf "%s-%s" $hostname .Values.dnsWildcard }}
{{- end }}

{{- define "control-plane-console.url" -}}
{{- printf "%s%s" "https://" (include "control-plane-console.hostname" .) }}
{{- end }}

{{- define "keycloak.realm" -}}
{{- printf "%s-%s" .Release.Namespace .Values.keycloakIntegration.realm }}
{{- end -}}
