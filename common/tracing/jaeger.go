package tracing

import (
	"strings"

	"coin-server/common/consulkv"
	"coin-server/common/proto/models"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/env"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type JaegerConfig struct {
	HttpAddr string `json:"httpAddr"`
	UdpAddr  string `json:"UdpAddr"`
}

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(cfg *JaegerConfig, name string, serverId values.ServerId) *tracesdk.TracerProvider {
	// Create the Jaeger exporter
	var err error
	var exp *jaeger.Exporter
	if cfg.HttpAddr != "" {
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.HttpAddr)))
	} else if cfg.UdpAddr != "" {
		list := strings.Split(strings.TrimSpace(cfg.UdpAddr), ":")
		exp, err = jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost(list[0]), jaeger.WithAgentPort(list[1])))
	}
	utils.Must(err)
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
			attribute.Int64("ID", serverId),
		)),
	)
	return tp
}

func Init(cnf *consulkv.Config, serverId values.ServerId, serverType models.ServerType) {
	if !env.OpenTracing() {
		return
	}
	// 优先使用环境变量里的jaeger地址，如果环境变量为空再使用consul里的地址
	cfg := &JaegerConfig{}
	envAddr := env.GetString(env.JEAGER_ADDR)
	envUdpAddr := env.GetString(env.JEAGER_UDP_ADDR)
	if envAddr != "" {
		cfg.HttpAddr = envAddr
	} else if envUdpAddr != "" {
		cfg.UdpAddr = envUdpAddr
	} else {
		err := cnf.Unmarshal("jaeger", cfg)
		utils.Must(err)
	}
	name := models.ServerType_name[int32(serverType)]
	provider := tracerProvider(cfg, name, serverId)
	otel.SetTracerProvider(provider)
}
