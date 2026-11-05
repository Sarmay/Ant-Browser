package proxy

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type directProxyBridgeSpec struct {
	Scheme   string
	Server   string
	Port     int
	Username string
	Password string
}

func RequiresLocalProxyBridgeForBrowser(src string) bool {
	spec, err := parseDirectProxyBridgeSpec(src)
	return err == nil && spec != nil
}

func buildDirectProxyBridgeOutbound(src string) (map[string]interface{}, bool, error) {
	spec, err := parseDirectProxyBridgeSpec(src)
	if err != nil {
		return nil, false, err
	}
	if spec == nil {
		return nil, false, nil
	}

	if spec.Scheme == "socks5" || spec.Scheme == "http" || spec.Scheme == "https" {
		return chainSocks5Outbound(chainSocks5Hop{
			Protocol: spec.Scheme,
			Server:   spec.Server,
			Port:     spec.Port,
			Username: spec.Username,
			Password: spec.Password,
			TLS:      spec.Scheme == "https",
		}, "proxy-out", ""), true, nil
	}

	return nil, false, nil
}

func parseDirectProxyBridgeSpec(src string) (*directProxyBridgeSpec, error) {
	raw := strings.TrimSpace(src)
	if raw == "" {
		return nil, nil
	}
	lowerRaw := strings.ToLower(raw)
	if !strings.HasPrefix(lowerRaw, "http://") && !strings.HasPrefix(lowerRaw, "https://") && !strings.HasPrefix(lowerRaw, "socks5://") {
		return nil, nil
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("代理地址解析失败: %w", err)
	}

	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	switch scheme {
	case "socks5", "http", "https":
		if parsed.User == nil {
			return nil, nil
		}
		username := strings.TrimSpace(parsed.User.Username())
		if username == "" {
			return nil, nil
		}
		server := rewriteHostProcessProxyHost(parsed.Hostname())
		if server == "" {
			return nil, fmt.Errorf("代理地址缺少主机名")
		}
		port, err := strconv.Atoi(parsed.Port())
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("代理端口无效")
		}
		password, _ := parsed.User.Password()
		return &directProxyBridgeSpec{
			Scheme:   scheme,
			Server:   server,
			Port:     port,
			Username: username,
			Password: password,
		}, nil
	default:
		return nil, nil
	}
}

// RewriteHostProcessProxyURL 把容器里才该使用的主机别名改回本机回环。
// Ant Browser 跑在宿主机上。host.docker.internal 在 macOS 上常被 TUN 解析成
// 198.18.0.0/15 假 IP，连上去时代理握手会被直接断开，浏览器就无法上网。
// 发布在宿主机上的 Resin 端口要通过 127.0.0.1 访问。
func RewriteHostProcessProxyURL(src string) string {
	raw := strings.TrimSpace(src)
	if raw == "" || strings.EqualFold(raw, "direct://") {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	host := rewriteHostProcessProxyHost(parsed.Hostname())
	if host == "" || strings.EqualFold(host, parsed.Hostname()) {
		return raw
	}
	port := parsed.Port()
	replacement := host
	if port != "" {
		replacement = net.JoinHostPort(host, port)
	}
	hostPort := parsed.Host
	idx := strings.LastIndex(raw, hostPort)
	if idx < 0 || hostPort == "" {
		return raw
	}
	return raw[:idx] + replacement + raw[idx+len(hostPort):]
}

func rewriteHostProcessProxyHost(host string) string {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "host.docker.internal", "gateway.docker.internal":
		return "127.0.0.1"
	default:
		return strings.TrimSpace(host)
	}
}
