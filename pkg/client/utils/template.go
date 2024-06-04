package utils

const CLIENT_TEMPLATE = `
# frpc.toml
serverAddr = "{{ .ServerAddress }}"
serverPort = {{ .ServerPort }}

{{ if eq .ServerAuthentication.Type 1 }}
auth.method = "token"
auth.token = "{{ .ServerAuthentication.Token }}"
{{ end }}

webServer.addr = "{{ .AdminAddress }}"
webServer.port = {{ .AdminPort }}
webServer.user = "{{ .AdminUsername }}"
webServer.password = "{{ .AdminPassword }}"

{{ range $upstream := .Upstreams }}

[{{ $upstream.Name }}]

{{ if eq $upstream.Type 1 }}
name = "{{ $upstream.Name }}"
type = "tcp"
localIP = "{{ $upstream.TCP.Host }}"
localPort = {{ $upstream.TCP.Port }}
remotePort = {{ $upstream.TCP.ServerPort }}

{{ if $upstream.TCP.ProxyProtocol }}
transport.proxyProtocolVersion = {{ $upstream.TCP.ProxyProtocol }}
{{ end }}


{{ if $upstream.TCP.HealthCheck }}
healthCheck.type = "tcp"
healthCheck.timeoutSeconds = $upstream.TCP.HealthCheck.TimeoutSeconds
healthCheck.maxFailed = $upstream.TCP.HealthCheck.MaxFailed
healthCheck.intervalSeconds = $upstream.TCP.HealthCheck.IntervalSeconds
{{ end }}

transport.useEncryption = true
{{ end }}

{{ if eq $upstream.Type 2 }}
name = "{{ $upstream.Name }}"
type = udp
localIP = "{{ $upstream.TCP.Host }}"
localPort = {{ $upstream.TCP.Port }}
remotePort = {{ $upstream.TCP.ServerPort }}
{{ end }}

{{ end }}
`
