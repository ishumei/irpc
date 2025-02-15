package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/bytedance/gopkg/util/gopool"
	json "github.com/bytedance/sonic"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/day253/krpc/conf"
	"github.com/day253/krpc/kclient"
	"github.com/day253/krpc/kserver"
	"github.com/day253/krpc/objects"
	"github.com/day253/krpc/protocols/arbiter/kitex_gen/com/shumei/service"
	"github.com/day253/krpc/zookeeper"
	"go.uber.org/ratelimit"
)

var (
	hostports   = flag.String("hostports", "", "hostports")
	metabase    = flag.String("metabase", "127.0.0.1:2181", "metabase")
	qps         = flag.Int("qps", 1, "qps")
	requestFile = flag.String("request_file", "request.json", "request file")
	serviceName = flag.String("service_name", "/test/models/boilerplate", "service name")
	timeout     = flag.Int("timeout", 1000, "timeout")
)

func main() {
	flag.Parse()
	flag.VisitAll(func(f *flag.Flag) {
		klog.Info(f.Name, ": ", f.Value)
	})
	client := kclient.MustNewArbiterClient(func() *kclient.SingleClientConf {
		c := &kserver.FrameConfig{}
		err := conf.LoadDefaultConf(c, "frame", "overwrite.yaml")
		klog.Info(objects.String(c))
		return &kclient.SingleClientConf{
			ResolverConf: kclient.ResolverConf{
				Hostports: func() []string {
					if *hostports != "" {
						return strings.Split(*hostports, ",")
					}
					return []string{}
				}(),
				Resolver: func() zookeeper.Conf {
					if err == nil {
						return c.Registry
					}
					return zookeeper.Conf{
						Metabase:  *metabase,
						TimeoutMs: 10000,
					}
				}(),
			},
			ClientConf: kclient.ClientConf{
				ServiceName: func() string {
					if err == nil {
						return c.ServiceName
					}
					return *serviceName
				}(),
				Retries:   0,
				TimeoutMs: *timeout,
			},
		}
	}())
	client.Start()
	defer func() {
		_ = client.Shutdown()
	}()
	r := &service.PredictRequest{}
	bs, err := os.ReadFile(*requestFile)
	if err != nil {
		klog.Error("Read request file error:", err)
	}
	if err := json.Unmarshal(bs, r); err != nil {
		klog.Error("Unmarshal request file error:", err)
	}
	rl := ratelimit.New(*qps)
	for {
		rl.Take()
		gopool.Go(func() {
			_, err := client.Predict(context.Background(), r)
			if err != nil {
				klog.Error(err)
			}
		})
	}
}
