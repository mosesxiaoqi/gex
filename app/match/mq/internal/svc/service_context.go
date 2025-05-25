package svc

import (
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/luxun9527/gex/app/match/mq/internal/config"
	"github.com/luxun9527/gex/app/match/mq/internal/dao/model"
	"github.com/luxun9527/gex/app/match/mq/internal/dao/query"
	"github.com/luxun9527/gex/app/order/rpc/orderservice"
	pulsarConfig "github.com/luxun9527/gex/common/pkg/pulsar"
	"github.com/luxun9527/gex/common/proto/define"
	// 一个用于 websocket 推送系统的 proto 包，提供 WebSocket 推送 RPC 客户端
	gpushPb "github.com/luxun9527/gpush/proto"
	logger "github.com/luxun9527/zlog"
	/*Go-Zero 框架提供的库：
		•	logx：日志输出
		•	redis：Redis 客户端
		•	zrpc：gRPC 客户端封装 */
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	MatchConsumer pulsar.Consumer // // 消费者，用于接收撮合相关消息
	Config        config.Config // 配置项，包括 pulsar、redis、gRPC 等
	OrderClient   orderservice.OrderService // 调用订单服务的 RPC 客户端
	Query         *query.Query // 数据库查询构造器（ORM/DAO）
	RedisClient   *redis.Redis // Redis 客户端
	WsClient      gpushPb.ProxyClient // WebSocket 推送 RPC 客户端
	MatchDataChan chan *model.MatchData // 撮合数据的通道，用于并发处理
	SymbolInfo    *define.SymbolInfo // 当前处理的交易对信息
}

// 它是一个构造函数，用于初始化整个服务运行所需的上下文 ServiceContext 结构体，
// 封装了各种依赖资源，包括日志、配置、消息队列（Pulsar）、Redis、数据库、WebSocket 客户端等
// c config.Config：这是读取配置文件反序列化得到的
func NewServiceContext(c config.Config) *ServiceContext {
	logger.InitDefaultLogger(&c.LoggerConfig)
	logx.SetWriter(logger.NewZapWriter(logger.GetZapLogger()))
	logx.DisableStat()

	var symbolInfo define.SymbolInfo
	// 把配置加载到 symbolInfo 结构体中，用于后续撮合逻辑处理
	define.InitSymbolConfig(define.EtcdSymbolPrefix+c.Symbol, c.SymbolEtcdConfig, &symbolInfo)
	logx.Infof("init symbol config symbolInfo %+v", &symbolInfo)

	client, err := c.PulsarConfig.BuildClient()
	if err != nil {
		logx.Severef("init pulsar client failed err %v", err)
	}
	topic := pulsarConfig.Topic{
		Tenant:    pulsarConfig.PublicTenant,
		Namespace: pulsarConfig.GexNamespace,
		Topic:     pulsarConfig.MatchResultTopic + "_" + c.Symbol,
	}
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic.BuildTopic(),
		SubscriptionName: pulsarConfig.MatchResultMatchSub,
		Type:             pulsar.Shared,
	})
	if err != nil {
		logx.Severef("init pulsar consumer failed %v", logger.ErrorField(err))
	}
	logx.Infof("init pulsar consumer success")
	sc := &ServiceContext{
		MatchConsumer: consumer,
		Config:        c,
		Query:         query.Use(c.GormConf.MustNewGormClient()),
		WsClient:      gpushPb.NewProxyClient(zrpc.MustNewClient(c.WsConf).Conn()),
		RedisClient:   redis.MustNewRedis(c.RedisConf),
		MatchDataChan: make(chan *model.MatchData),
		SymbolInfo:    &symbolInfo,
	}
	return sc
}
