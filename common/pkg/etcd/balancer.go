package etcd

// gRPC 允许你自定义负载均衡策略。
// base 包封装了一套基础功能（比如子连接管理、状态传播、picker 注册等），
// 你只需要关注自己的策略逻辑，比如你要按 symbol 路由、按权重选择连接等
import (
	"github.com/spf13/cast"
	"google.golang.org/grpc/balancer"
	// base 提供了一套基础设施，用于快速构建自定义的 gRPC 负载均衡器（balancer）
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

//基于交易对负载均衡
//场景 撮合，订单，行情 分交易对，api 服务对这些交易对建立连接后

//key klineRpc/IKUN_USDT/xxxxx

func init() {
	balancer.Register(newSymbolBalancerBuilder())
}

// 定义了一个错误变量
var (
	NotAvailableConn = status.Error(codes.Unavailable, "no available connection")
)

// 一个常量
const SymbolLB = "symbol_lb"

// 自定义 Picker
type symbolPicker struct {
	subConns map[string][]balancer.SubConn // 连接列表
}

func (p *symbolPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	// 如果没有可用的连接，返回错误
	if len(p.subConns) == 0 {
		return balancer.PickResult{}, NotAvailableConn
	}
	md, ok := metadata.FromIncomingContext(info.Ctx)
	if !ok {
		return balancer.PickResult{}, NotAvailableConn
	}
	symbol := md.Get("symbol")[0]
	conns, ok := p.subConns[symbol]
	if !ok || len(conns) == 0 {
		return balancer.PickResult{}, NotAvailableConn
	}
	index := time.Now().UnixNano() % int64(len(conns))
	return balancer.PickResult{SubConn: conns[index]}, nil
}

// 负载均衡器构建器
type symbolPickerBuilder struct {
	weightConfig map[string]int32
}

func (wp *symbolPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	var p = map[string][]balancer.SubConn{}
	for sc, addr := range info.ReadySCs {
		symbolData, ok := addr.Address.Metadata.(map[string]interface{})
		if !ok {
			continue
		}
		symbol := cast.ToString(symbolData["symbol"])
		if symbol == "" {
			continue
		}
		conns, ok := p[symbol]
		if !ok {
			conns = make([]balancer.SubConn, 0, 1)
		}
		conns = append(conns, sc)
		p[symbol] = conns
	}

	return &symbolPicker{
		subConns: p,
	}
}

// 自定义负载均衡
func newSymbolBalancerBuilder() balancer.Builder {
	// &symbolPickerBuilder{}表示创建一个 symbolPickerBuilder 类型的结构体实例，并返回它的指针
	// symbolPickerBuilder{}创建结构体实例
	// &表示取地址
	// base.NewBalancerBuilder用来快速构建一个新的 balancer.Builder
	// config：是否需要 subConn 状态平滑处理等（通常传 base.Config{HealthCheck: true}）
	// 然后把这个 builder 注册进去 balancer.Register(builder)
	return base.NewBalancerBuilder(SymbolLB, &symbolPickerBuilder{}, base.Config{HealthCheck: true})
}
