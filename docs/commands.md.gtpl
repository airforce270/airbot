# Commands

All commands assume the `$` prefix, but the prefix is configurable
per-channel (see [$setprefix](#setprefix)).

Some commands include parameters.

If the parameter is wrapped in `<angle brackets>`, it's a **required** parameter.

If it's wrapped in `[square brackets]`, it's an **optional** parameter.

{{- range $groupName, $commands := . }}

## {{ $groupName }}

{{- range $commands }}

### ${{ .Name }}
{{ if .Help }}
- {{ .Help }}
{{- end }}
- > Usage: `{{ formatUsage . }}`
{{- if and .Permission .Permission.IsElevated }}
- > Minimum permission level: `{{ .Permission.Name }}`
{{- end }}
{{- if .ChannelCooldown }}
- > Per-channel cooldown: `{{ .ChannelCooldown }}`
{{- end }}
{{- if .UserCooldown }}
- > Per-user cooldown: `{{ .UserCooldown }}`
{{- end }}
{{- if .Aliases }}
- > Aliases: {{ formatAliases .Aliases }}
{{- end }}

{{- end }}{{/* command */}}

{{- end }}{{/* groups */}}
