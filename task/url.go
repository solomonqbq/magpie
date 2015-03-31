package task

import (
	"errors"
	"github.com/xeniumd-china/xeniumd-monitor/common"
	"strings"
)

type URL struct {
	Protocol   string
	Username   string
	Password   string
	Host       string
	Port       uint16
	Path       string
	Parameters map[string]string
}

func (this *URL) GetParameter(key string, defaultValue string) string {
	value, exist := this.Parameters[key]
	if exist {
		return value
	} else {
		return defaultValue
	}
}

func Parse(text string) (*URL, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text is null")
	}
	var protocol string
	var username string
	var password string
	var host string
	var port uint16
	var path string
	var parameters map[string]string

	// cloud://user:password@jss.360buy.com/mq?timeout=60000
	// file:/path/to/file.txt
	// zk://10.10.10.10:2181,10.10.10.11:2181/?retryTimes=3
	// failover://(zk://10.10.10.10:2181,10.10.10.11:2181;zk://20.10.10.10:2181,20.10.10.11:2181)?interval=1000
	j := 0
	i := strings.Index(text, ")")
	if i >= 0 {
		i = strings.Index(text[i:], "?")
	} else {
		i = strings.Index(text, "?")
	}
	if i >= 0 {
		if i < len(text)-1 {
			parts := strings.Split(text[i+1:], "&")
			parameters = make(map[string]string)
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					j = strings.Index(part, "=")
					if j > 0 {
						if j == len(part)-1 {
							parameters[part[:j]] = ""
						} else {
							parameters[part[:j]] = part[j+1:]
						}
					} else if j == -1 {
						parameters[part] = part
					}
				}
			}
		}
		text = text[:i]
	}
	i = strings.Index(text, "://")
	if i > 0 {
		protocol = text[:i]
		text = text[i+3:]
	} else if i < 0 {
		// case: file:/path/to/file.txt
		i = strings.Index(text, ":/")
		if i > 0 {
			protocol = text[:i]
			// 保留路径符号“/”
			text = text[i+1:]
		}
	}
	if protocol == "" {
		return nil, errors.New("url missing protocol: " + text)
	}

	i = strings.LastIndex(text, ")")
	if i >= 0 {
		i = strings.Index(text[i:], "/")
	} else {
		i = strings.Index(text, "/")
	}
	if i >= 0 {
		path = text[i+1:]
		text = text[:i]
	}
	i = strings.Index(text, "(")
	if i >= 0 {
		j = strings.LastIndex(text, ")")
		if j >= 0 {
			text = text[i+1 : j]
		} else {
			text = text[i+1:]
		}
	} else {
		i = strings.Index(text, "@")
		if i >= 0 {
			username = text[:i]
			j = strings.Index(username, ":")
			if j >= 0 {
				password = username[j+1:]
				username = username[0:j]
			}
			text = text[i+1:]
		}
		values := strings.Split(text, ":")
		if len(values) == 2 {
			// 排除zookeeper://192.168.1.2:2181,192.168.1.3:2181
			port = common.ParseUint16(values[1])
			text = values[0]
		}
	}
	if text != "" {
		host = text
	}
	return NewURL(protocol, username, password, host, port, path, parameters)
}

func NewURL(protocol string, username string, password string, host string, port uint16, path string, parameters map[string]string) (*URL, error) {
	if username == "" && password != "" {
		return nil, errors.New("Invalid url, password without username!")
	}
	url := new(URL)
	url.Protocol = protocol
	url.Username = username
	url.Password = password
	url.Host = host
	url.Port = port
	// trim the beginning "/"
	for path != "" && strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	url.Path = path
	url.Parameters = make(map[string]string)
	for key, value := range parameters {
		url.Parameters[key] = value
	}
	return url, nil
}
