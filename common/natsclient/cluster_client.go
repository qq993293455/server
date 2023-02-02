package natsclient

import (
	"context"
	"encoding/binary"
	"hash/crc32"
	"math/rand"
	"time"

	"coin-server/common/bytespool"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/env"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/gogo/protobuf/proto"
)

type ClusterClient struct {
	natsClients []*NatsClient
}

var natsTracer = otel.Tracer("NatsClient")

func NewClusterClient(serverType models.ServerType, serverId int64, urls []string, log *logger.Logger) *ClusterClient {
	clusterClient := &ClusterClient{}
	clusterClient.natsClients = make([]*NatsClient, 0, len(urls))
	for _, url := range urls {
		natsClient := NewNatsClient(serverType, serverId, url, log)
		clusterClient.natsClients = append(clusterClient.natsClients, natsClient)
	}
	return clusterClient
}

func (this_ *ClusterClient) Close() {
	for _, natsClient := range this_.natsClients {
		natsClient.Close()
	}
}

func (this_ *ClusterClient) PublishCtx(c *ctx.Context, serverId int64, msg proto.Message) *errmsg.ErrMsg {
	if env.OpenTracing() {
		var span trace.Span
		c.Context, span = natsTracer.Start(c.Context, "PublishCtx")
		defer span.End()
		span.SetAttributes(attribute.String("name", proto.MessageName(msg)))
	}
	nc := this_.getNatsClientWithCtx(c, serverId)
	return nc.Publish(serverId, c.ServerHeader, msg)
}

func (this_ *ClusterClient) Publish(serverId int64, h *models.ServerHeader, msg proto.Message) *errmsg.ErrMsg {
	if env.OpenTracing() {
		var span trace.Span
		_, span = natsTracer.Start(context.Background(), "Publish")
		defer span.End()
		span.SetAttributes(attribute.String("name", proto.MessageName(msg)))
	}
	nc := this_.getNatsClientWithHeader(h, serverId)
	return nc.Publish(serverId, h, msg)
}

func (this_ *ClusterClient) PublishRawData(serverId int64, h *models.ServerHeader, msgName string, msgData []byte) *errmsg.ErrMsg {
	if env.OpenTracing() {
		var span trace.Span
		_, span = natsTracer.Start(context.Background(), "PublishRawData")
		defer span.End()
		span.SetAttributes(attribute.String("name", msgName))
	}
	nc := this_.getNatsClientWithHeader(h, serverId)
	return nc.PublishRawData(serverId, h, msgName, msgData)
}

func (this_ *ClusterClient) Shutdown() {
	for _, natsClient := range this_.natsClients {
		natsClient.Shutdown()
	}
}

func (this_ *ClusterClient) Subscribe(subj string, h nats.MsgHandler) {
	for _, natsClient := range this_.natsClients {
		natsClient.Subscribe(subj, h)
	}
}
func (this_ *ClusterClient) RequestWithOut(c *ctx.Context, serverId int64, msg proto.Message, out proto.Message, timeout ...time.Duration) *errmsg.ErrMsg {
	if env.OpenTracing() {
		var span trace.Span
		_, span = natsTracer.Start(c.Context, "RequestWithOut")
		defer span.End()
		span.SetAttributes(attribute.String("name", proto.MessageName(out)))
	}
	nc := this_.getNatsClientWithCtx(c, serverId)
	return nc.RequestWithOut(c, serverId, msg, out, timeout...)
}

func (this_ *ClusterClient) RequestProto(serverId int64, header *models.ServerHeader, msg proto.Message) ([]byte, *errmsg.ErrMsg) {
	if env.OpenTracing() {
		var span trace.Span
		_, span = natsTracer.Start(context.Background(), "RequestProto")
		defer span.End()
		span.SetAttributes(attribute.String("name", proto.MessageName(msg)))
	}
	nc := this_.getNatsClientWithHeader(header, serverId)
	return nc.RequestProto(serverId, header, msg)
}

func (this_ *ClusterClient) RequestData(serverId int64, header *models.ServerHeader, msgName string, data []byte) ([]byte, *errmsg.ErrMsg) {
	if env.OpenTracing() {
		var span trace.Span
		_, span = natsTracer.Start(context.Background(), "RequestData")
		defer span.End()
		span.SetAttributes(attribute.String("name", msgName))
	}
	nc := this_.getNatsClientWithHeader(header, serverId)
	return nc.RequestData(serverId, header, msgName, data)
}

func (this_ *ClusterClient) Request(ctx *ctx.Context, serverId values.ServerId, msg proto.Message) (*models.Resp, *errmsg.ErrMsg) {
	if env.OpenTracing() {
		var span trace.Span
		ctx.Context, span = natsTracer.Start(ctx.Context, "Request")
		defer span.End()
		span.SetAttributes(attribute.String("name", proto.MessageName(msg)))
	}
	nc := this_.getNatsClientWithCtx(ctx, serverId)
	return nc.Request(ctx, serverId, msg)
}

func (this_ *ClusterClient) RequestWithHeader(ctx *ctx.Context, serverId values.ServerId, header *models.ServerHeader, msg proto.Message) (*models.Resp, *errmsg.ErrMsg) {
	if env.OpenTracing() {
		var span trace.Span
		ctx.Context, span = natsTracer.Start(ctx.Context, "RequestWithHeader")
		defer span.End()
		span.SetAttributes(attribute.String("name", proto.MessageName(msg)))
	}
	nc := this_.getNatsClientWithHeader(header, serverId)
	return nc.RequestWithHeader(serverId, header, msg)
}

func (this_ *ClusterClient) RequestWithHeaderOut(ctx *ctx.Context, serverId values.ServerId, header *models.ServerHeader, msg proto.Message, out proto.Message) *errmsg.ErrMsg {
	if env.OpenTracing() {
		var span trace.Span
		ctx.Context, span = natsTracer.Start(ctx.Context, "RequestWithHeader")
		defer span.End()
		span.SetAttributes(attribute.String("name", proto.MessageName(msg)))
	}
	nc := this_.getNatsClientWithHeader(header, serverId)
	return nc.RequestWithHeaderOut(serverId, header, msg, out)
}

func (this_ *ClusterClient) SubscribeHandler(subj string, f func(ctx *ctx.Context)) {
	for _, natsClient := range this_.natsClients {
		natsClient.SubscribeHandler(subj, f)
	}
}

func (this_ *ClusterClient) UnSub(subj string) {
	for _, natsClient := range this_.natsClients {
		natsClient.UnSub(subj)
	}
}

func (this_ *ClusterClient) SubscribeBroadcast(subj string, f func(ctx *ctx.Context)) {
	for _, natsClient := range this_.natsClients {
		natsClient.SubscribeBroadcast(subj, f)
	}
}

func (this_ *ClusterClient) getNatsClientWithHeader(sh *models.ServerHeader, serverId int64) *NatsClient {
	nsl := len(this_.natsClients)
	if nsl == 0 {
		panic("nats client is empty")
	}
	if nsl == 1 {
		return this_.natsClients[0]
	}
	var temp []byte
	if sh != nil {

		if sh.RoleId != "" {
			temp = utils.StringToBytes(sh.RoleId)
			return this_.natsClients[crc32.ChecksumIEEE(temp)%uint32(nsl)]
		}
		if sh.UserId != "" {
			temp = utils.StringToBytes(sh.UserId)
			return this_.natsClients[crc32.ChecksumIEEE(temp)%uint32(nsl)]
		}
		if sh.ServerId > 0 {
			temp = bytespool.GetSample(8)
			binary.LittleEndian.PutUint64(temp, uint64(sh.ServerId))
			return this_.natsClients[crc32.ChecksumIEEE(temp)%uint32(nsl)]
		}

	}
	if serverId > 0 {
		temp = bytespool.GetSample(8)
		defer bytespool.PutSample(temp)
		binary.LittleEndian.PutUint64(temp, uint64(serverId))
		return this_.natsClients[crc32.ChecksumIEEE(temp)%uint32(nsl)]
	}
	return this_.natsClients[rand.Intn(nsl)]
}

func (this_ *ClusterClient) getNatsClientWithCtx(c *ctx.Context, serverId values.ServerId) *NatsClient {
	return this_.getNatsClientWithHeader(c.ServerHeader, serverId)
}
