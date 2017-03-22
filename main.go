package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/eSailors/go-datadog"
	"github.com/gorilla/handlers"
	"github.com/jessevdk/go-flags"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/rcrowley/go-metrics"
	"github.com/x-cray/logrus-prefixed-formatter"
	"gopkg.in/tylerb/graceful.v1"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"fmt"
	"github.com/mopsalarm/go-remote-config/config"
	"github.com/mopsalarm/go-remote-config/restapi"
)

func main() {
	metrics.DefaultRegistry = metrics.NewPrefixedRegistry("pr0gramm.config.")

	var err error

	var opts struct {
		Http struct {
			Listen     string `long:"listen" default:":3000" value-name:"ADDR" description:"Address the http server will use to listen for client connections."`
			AddressLog string `long:"access-log" default:"-" value-name:"FILE" description:"File to write an http access log to. Leave empty to disable access logging."`
		} `namespace:"http" group:"HTTP server options"`

		Datadog struct {
			Apikey string `long:"apikey" value-name:"KEY" description:"Datadog api key. Set this to enable reporting or leave it empty to skip datadog reporting."`
			Tags   string `long:"tags" value-name:"TAGS" description:"Comma separated list of tags to add to datadog metrics."`
		} `namespace:"datadog" group:"Datadog metrics reporting"`

		Config        string `long:"config" default:"config.json" description:"Path to json config file containing the rules for this service. The file will be overwritten if new rules are posted."`
		AdminPassword string `long:"admin-password" value-name:"PWD" default:"admin" description:"Admin password to secure rule updates with."`

		Verbose bool `short:"v" long:"verbose" description:"Enable verbose logging"`
	}

	parser := flags.NewParser(&opts, flags.Default)
	parser.NamespaceDelimiter = "-"
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	// enable prefix logger for logrus
	log.SetFormatter(&prefixed.TextFormatter{})

	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
		log.Debug("Verbose logging enabled")
	}

	initializeMetrics()
	initializeDatadogReporting(opts.Datadog.Apikey, strings.FieldsFunc(opts.Datadog.Tags, isComma))

	router := httprouter.New()

	rules, err := config.Load(opts.Config)
	fatalOnError(err, "Could not load config from file at %s", err)

	restapi.Setup(router, opts.AdminPassword, opts.Config, rules)

	err = httpListen(opts.Http.Listen, opts.Http.AddressLog, router)
	fatalOnError(err, "Could not start http server")

	log.Info("Bye.")

}

func fatalOnError(err error, reason string, args ...interface{}) {
	if err != nil {
		log.Fatalf("%s: %s", fmt.Sprintf(reason, args), err)
	}
}

func initializeMetrics() {
	metrics.RegisterRuntimeMemStats(metrics.DefaultRegistry)
	go metrics.CaptureRuntimeMemStats(metrics.DefaultRegistry, 60*time.Second)
}

// Start metrics reporting to datadog. This starts a reporter that sends the
// applications metrics once per minute to datadog if you provide a valid api key.
func initializeDatadogReporting(apikey string, tags []string) {
	if apikey != "" {
		hostname, _ := os.Hostname()

		log.Debugf("Starting datadog reporting on hostname '%s' with tags: %s",
			hostname, strings.Join(tags, ", "))

		client := datadog.New(hostname, apikey)
		go datadog.Reporter(client, metrics.DefaultRegistry, tags).Start(60 * time.Second)
	}
}

// Returns true if the given rune is equal to a comma.
func isComma(ch rune) bool {
	return ch == ','
}

func httpListen(addr string, accessLog string, handler http.Handler) error {
	// open access-log file or logger
	if accessLog != "" && accessLog != "/dev/null" {
		var writer io.WriteCloser
		if accessLog == "-" {
			writer = log.StandardLogger().Writer()
		} else {
			log.Debugf("Opening access-log file at %s", accessLog)
			var err error
			writer, err = os.OpenFile(accessLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return errors.WithMessage(err, "Could not open access-log at "+accessLog)
			}
		}

		defer writer.Close()
		handler = handlers.LoggingHandler(writer, handler)
	}

	// catch and log panics from the http handlers
	handler = handlers.RecoveryHandler(
		handlers.PrintRecoveryStack(true),
		handlers.RecoveryLogger(log.StandardLogger().WithField("prefix", "httpd")),
	)(handler)

	log.Infof("Starting http server on %s now.", addr)
	server := &graceful.Server{
		Timeout: 10 * time.Second,
		LogFunc: log.StandardLogger().WithField("prefix", "httpd").Warnf,
		Server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}

	return server.ListenAndServe()
}
