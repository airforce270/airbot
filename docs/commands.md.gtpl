# Commands

All commands assume the `$` prefix, but note that the prefix is configurable
per-channel (in `config.toml`).
To find out what the prefix is in a channel, ask `what's airbot's prefix?`
in a chat.

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
{{- if .Usage }}
- > Usage: `{{ .Usage }}`
{{- end }}
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
