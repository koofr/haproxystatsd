package haproxystatsd

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"regexp"

	"github.com/jeromer/syslogparser"
	"gopkg.in/mcuadros/go-syslog.v2"
)

var (
	// this MUST contain some named groups
	DefaultLogPattern = regexp.MustCompile(`.* \[.*\] (?P<frontend_name>.*?) (?P<backend_name>.*?)/(?P<server_name>.*?) (?P<Tq>\d+)/(?P<Tw>\d+)/(?P<Tc>\d+)/(?P<Tr>\d+)/(?P<Tt>\d+) (?P<status_code>\d+) \d+ .*? .*? .* (?P<actconn>\d+)/(?P<feconn>\d+)/(?P<beconn>\d+)/(?P<srv_conn>\d+)/(?P<retries>\d+) (?P<srv_queue>\d+)/(?P<backend_queue>\d+)`)
	// Use names from named groups in regex to construct bucket prefix
	DefaultBucketTemplate = "{{.TAG}}.{{.frontend_name}}.{{.backend_name}}.{{.server_name}}"
)

type statsdSender interface {
	Send([]string)
}

type Config struct {
	StatsdAddr     string
	SyslogBindAddr string
	NodeTag        string
	LogPattern     string
	BucketTemplate string
	DryRun         bool
}

type HaproxyStatsd struct {
	srv        *syslog.Server
	sender     statsdSender
	logPattern *regexp.Regexp
	bucketTpl  *template.Template
	prefixKeys []string
	nodeTag    string
}

func New(cfg *Config) (hs *HaproxyStatsd, err error) {

	hs = &HaproxyStatsd{
		srv: syslog.NewServer(),
	}

	// set defaults
	if cfg.LogPattern == "" {
		hs.logPattern = DefaultLogPattern
	}
	if cfg.BucketTemplate == "" {
		cfg.BucketTemplate = DefaultBucketTemplate
	}

	if cfg.NodeTag == "" {
		hs.nodeTag, _ = os.Hostname()
	} else {
		hs.nodeTag = cfg.NodeTag
	}

	// compile stuff
	if hs.logPattern, err = regexp.Compile(cfg.LogPattern); err != nil {
		return
	}

	if hs.bucketTpl, err = template.New("bucket").Parse(cfg.BucketTemplate); err != nil {
		return
	}

	// identify template keys
	hs.prefixKeys = parsePrefixKeys(cfg.BucketTemplate)

	hs.srv.SetFormat(syslog.Automatic)
	hs.srv.ListenUDP(cfg.SyslogBindAddr)
	hs.srv.SetHandler(hs)
	return
}

func (h *HaproxyStatsd) Boot() (err error) {
	err = h.srv.Boot()
	return
}

func (h *HaproxyStatsd) Wait() {
	h.srv.Wait()
}

func (h *HaproxyStatsd) Handle(parts syslogparser.LogParts, woot int64, err error) {

	kv := make(map[string]string)
	kvdata := make(map[string]string)

	switch content := parts["content"].(type) {
	case string:
		groups := DefaultLogPattern.SubexpNames()
		matches := DefaultLogPattern.FindStringSubmatch(content)
		for i, n := range matches {
			switch {
			case groups[i] == "":
			case sliceContains(h.prefixKeys, groups[i]):
				kv[groups[i]] = n
			default:
				kvdata[groups[i]] = n
			}
		}
	}

	kv["TAG"] = h.nodeTag

	msgs := h.createStatsdMessages(kv, kvdata)
	h.sender.Send(msgs)
}

func (h *HaproxyStatsd) createStatsdMessages(kv map[string]string, kvdata map[string]string) (msgs []string) {
	prefix := h.createStatsdPrefix(kv)
	for k, v := range kvdata {
		msg := ""
		if k == "status_code" {
			suffix := statusCode2Class(parseInt(v))
			msg = fmt.Sprintf("%s.%s:%s|c", prefix, suffix, v)
		} else {
			msg = fmt.Sprintf("%s.%s:%s|g", prefix, k, v)
		}
		msgs = append(msgs, msg)
	}
	return
}

func (h *HaproxyStatsd) createStatsdPrefix(kv map[string]string) string {
	buf := new(bytes.Buffer)
	h.bucketTpl.Execute(buf, kv)
	return buf.String()
}

func statusCode2Class(code int) string {
	switch {
	case code > 99 && code < 200:
		return "1xx"
	case code > 199 && code < 300:
		return "2xx"
	case code > 299 && code < 400:
		return "3xx"
	case code > 399 && code < 500:
		return "4xx"
	case code > 499 && code < 600:
		return "5xx"
	default:
		return "xxx"
	}
}

func parseInt(in string) int {
	i64, _ := strconv.ParseInt(strings.TrimSpace(in), 10, 32)
	return int(i64)
}

func parsePrefixKeys(tpl string) []string {
	ret := make([]string, 0)
	re := regexp.MustCompile(`\{\{\.(.*?)\}\}`)

	for _, match := range re.FindAllStringSubmatch(tpl, -1) {
		ret = append(ret, match[1])
	}
	return ret

}

func sliceContains(in []string, test string) bool {
	for _, x := range in {
		if x == test {
			return true
		}
	}
	return false
}
