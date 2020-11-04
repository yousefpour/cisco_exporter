package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

const version string = "0.1"

var (
	showVersion       = flag.Bool("version", false, "Print version information.")
	listenAddress     = flag.String("web.listen-address", ":9362", "Address on which to expose metrics and web interface.")
	metricsPath       = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	sshHosts          = flag.String("ssh.targets", "", "SSH Hosts to scrape")
	sshUsername       = flag.String("ssh.user", "cisco_exporter", "Username to use for SSH connection")
	sshKeyFile        = flag.String("ssh.keyfile", "", "Key file to use for SSH connection")
	sshTimeout        = flag.Int("ssh.timeout", 5, "Timeout to use for SSH connection")
	sshBatchSize      = flag.Int("ssh.batch-size", 10000, "The SSH response batch size")
	debug             = flag.Bool("debug", false, "Show verbose debug output in log")
	legacyCiphers     = flag.Bool("legacy.ciphers", false, "Allow legacy CBC ciphers")
	bgpEnabled        = flag.Bool("bgp.enabled", true, "Scrape bgp metrics")
	environmetEnabled = flag.Bool("environment.enabled", true, "Scrape environment metrics")
	factsEnabled      = flag.Bool("facts.enabled", true, "Scrape system metrics")
	interfacesEnabled = flag.Bool("interfaces.enabled", true, "Scrape interface metrics")
	opticsEnabled     = flag.Bool("optics.enabled", true, "Scrape optic metrics")
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: cisco_exporter [ ... ]\n\nParameters:")
		fmt.Println("cisco non-admin users - 0.0.2")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	startServer()
}

func printVersion() {
	fmt.Println("cisco_exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Martin Poppen")
	fmt.Println("Metric exporter for switches and routers running cisco IOS/NX-OS/IOS-XE")
}

func startServer() {
	log.Infof("Starting Cisco exporter (Version: %s)\n", version)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Cisco Exporter (Version ` + version + `)</title></head>
			<body>
			<h1>Cisco Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<h2>More information:</h2>
			<p><a href="https://github.com/lwlcom/cisco_exporter">github.com/lwlcom/cisco_exporter</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc(*metricsPath, handleMetricsRequest)

	log.Infof("Listening for %s on %s\n", *metricsPath, *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	reg := prometheus.NewRegistry()

	targets := strings.Split(*sshHosts, ",")
	c := newCiscoCollector(targets)
	reg.MustRegister(c)

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      log.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)
}
