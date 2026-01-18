#
# Software Name : free5gc-helm
# SPDX-FileCopyrightText: Copyright (c) 2021 Orange
# SPDX-License-Identifier: Apache-2.0
#
# This software is distributed under the Apache License 2.0,
# the text of which is available at https://github.com/Orange-OpenSource/towards5gs-helm/blob/main/LICENSE
# or see the "LICENSE" file for more details.
#
# Author: Adapted for NWDAF
# Software description: An open-source project providing Helm charts to deploy 5G components (Core + RAN) on top of Kubernetes
#
{{/*
Expand the name of the chart.
*/}}
{{- define "free5gc-nwdaf.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "free5gc-nwdaf.fullname" -}}
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
{{- define "free5gc-nwdaf.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "free5gc-nwdaf.labels" -}}
helm.sh/chart: {{ include "free5gc-nwdaf.chart" . }}
{{ include "free5gc-nwdaf.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "free5gc-nwdaf.selectorLabels" -}}
app.kubernetes.io/name: {{ include "free5gc-nwdaf.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
NWDAF Pod Annotations
*/}}
{{- define "free5gc-nwdaf.nwdafAnnotations" -}}
{{- with .Values.nwdaf }}
{{- if .podAnnotations }}
{{- toYaml .podAnnotations }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "free5gc-nwdaf.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "free5gc-nwdaf.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the NWDAF image name
*/}}
{{- define "free5gc-nwdaf.image" -}}
{{- $tag := .Values.nwdaf.image.tag | default .Chart.AppVersion -}}
{{- printf "%s:%s" .Values.nwdaf.image.name $tag -}}
{{- end }}

{{/*
Return the init container image name
*/}}
{{- define "free5gc-nwdaf.initImage" -}}
{{- printf "%s/%s:%s" .Values.initcontainers.curl.registry .Values.initcontainers.curl.image .Values.initcontainers.curl.tag -}}
{{- end }}
