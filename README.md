srvd
----

`srvd` is a configuration management tool using [DNS SRV record](https://en.wikipedia.org/wiki/SRV_record) like [confd](https://github.com/kelseyhightower/confd).

[![Build Status](https://travis-ci.org/winebarrel/srvd.svg?branch=master)](https://travis-ci.org/winebarrel/srvd)

## Usage

```sh
srvd -config srvd.toml
```

## Configuration example

```toml
src = "/etc/haproxy/haproxy.cfg.tmpl"
dest = "/etc/haproxy/haproxy.cfg"
domains = ["_http._tcp.example.com"]
reload_cmd = "/bin/systemctl reload haproxy.service"
check_cmd = "/usr/sbin/haproxy -c -V -f {{ .src }}"
interval = 1
timeout = 3
#resolv_conf = "/etc/resolv.conf"
cooldown = 60
#status_port = 8080
```

## Template example

```
backend nodes
  mode tcp
  {{ $srvs := index .domains "_http._tcp.example.com" }}
  # see https://godoc.org/github.com/miekg/dns#SRV
  {{ range $srvs }}
  server {{ .Target }} {{ .Target }}:{{ .Port }}
  {{ end }}
```

## Check status

```sh
$ curl localhost:8080/status
{"LastUpdate":"2018-08-02T23:38:25.647297201+09:00","Ok":true}
```
