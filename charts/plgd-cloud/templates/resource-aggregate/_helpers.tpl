{{- define  "plgd-cloud.resourceaggregate.image" -}}
    {{- $registryName := .Values.resourceaggregate.image.registry | default "" -}}
    {{- $repositoryName := .Values.resourceaggregate.image.repository -}}
    {{- $tag := .Values.resourceaggregate.image.tag | default .Chart.AppVersion | toString -}}
    {{- printf "%s%s:%s" $registryName $repositoryName  $tag -}}
{{- end -}}

{{- define  "plgd-cloud.resourceaggregate.configSecretName" -}}
    {{- $fullName :=  include "plgd-cloud.resourceaggregate.fullname" . -}}
    {{- printf "%s-cfg" $fullName }}
{{- end -}}

{{- define "plgd-cloud.resourceaggregate.createServiceCertByCm" }}
    {{- $serviceTls := .Values.resourceaggregate.apis.grpc.tls.certFile }}
    {{- if $serviceTls }}
    {{- printf "false" }}
    {{- else }}
    {{- printf "true" }}
    {{- end }}
{{- end }}

{{- define "plgd-cloud.resourceaggregate.serviceCertName" -}}
  {{- $fullName := include "plgd-cloud.resourceaggregate.fullname" . -}}
  {{- printf "%s-crt" $fullName -}}
{{- end }}

{{- define "plgd-cloud.resourceaggregate.selectorLabels" -}}
app.kubernetes.io/name: {{ .Values.resourceaggregate.name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}