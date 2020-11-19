package main

import (
	"flag"
	"fmt"
	"kitdemo/addsvc/pkg/addendpoint"
	"kitdemo/addsvc/pkg/addservice"
	"kitdemo/addsvc/pkg/addtransport"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	fs := flag.NewFlagSet("addsvc", flag.ExitOnError)
	var (
		//debugAddr = fs.String("debug-addr", ":8080", "Debug and metrics listen address")
		grpcAddr = fs.String("grpc-addr", ":8082", "gRPC Listen Address")
	)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags] ")
	fs.Parse(os.Args[1:])

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	var ints, chars metrics.Counter
	{
		ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "integers_summed",
			Help:      "Sum 方法求出的和的总数",
		}, []string{})
		chars = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "characters_concatednated",
			Help:      "Concat 方法连接的字符串的总长度",
		}, []string{})
	}
	var duration metrics.Histogram
	{
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "request_duration_seconds",
			Help:      "请求的总耗时",
		}, []string{"method", "success"})
	}
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	var (
		service    = addservice.New(logger, ints, chars)
		endpoints  = addendpoint.New(service, logger, duration)
		grpcServer = addtransport.NewGRPCServer(endpoints, logger)
	)

}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "Usage\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
