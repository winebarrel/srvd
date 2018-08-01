srvd
----

`srvd` is a configuration management tool using [DNS SRV record](https://en.wikipedia.org/wiki/SRV_record) like [confd](https://github.com/kelseyhightower/confd).

## Usage

```sh
srvd -config srvd.toml
```

## Configuration example

```toml
src = "/etc/haproxy/haproxy.cfg.tmpl"
dest = "/etc/haproxy/haproxy.cfg.tmpl"
domain = "_http._tcp.example.com"
reload_cmd = "/usr/local/sbin/haproxy -c -V -f {{ .src }}"
check_cmd = "/bin/systemctl reload haproxy.service"
interval = 1
timeout = 3
#resolv_conf = "/etc/resolv.conf"
cooldown = 60
```

## Template example

```
backend nodes
  mode tcp
  # see https://godoc.org/github.com/miekg/dns#SRV
  {{ range .srvs }}
  server {{ .Target }} {{ .Target }}:{{ .Port }}
  {{ end }}
```
