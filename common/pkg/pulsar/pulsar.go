package pulsar

import (
	"github.com/apache/pulsar-client-go/pulsar"
	pulsarLog "github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)


// 创建了一盒=个结构体，结构体字段是切片类型
type PulsarConfig struct {
	//字段可以存储集群的多个节点地址
	// 标签表示在进行序列化时，json和yaml文件中Hosts的键名
	Hosts []string `json:"hosts" yaml:"Hosts"`
}
//这是一个方法
// 对PulsarConfig实现对方法
// Pulsar 是一个分布式消息队列系统，
// 这个方法的目的是初始化一个客户端，用于与 Pulsar 集群通信
func (pc PulsarConfig) BuildClient() (pulsar.Client, error) {
	// 创建一个标准的 `logrus` 日志记录器。
	logger := logrus.StandardLogger()
	// 设置日志级别
	logger.Level = logrus.WarnLevel
	// make智能用于创建切片、map、channel（通道）
	// make相比new除了分配内存还会进行初始化
	// make返回初始化的值，new返回初始化后的指针
	addr := make([]string, 0, len(pc.Hosts))
	for _, v := range pc.Hosts {
		// 随着循环更新切片addr，存储完整的集群地址
		addr = append(addr, "pulsar://"+v)
	}
	// Join将addr里保存的多喝主机地址用 分隔符','连接起来
	// 例如：`[]string{"pulsar://127.0.0.1:6650", "pulsar://127.0.0.2:6650"}`。
	// - **`url`**：
	//  - 是一个字符串，表示多个主机地址的组合。
	//  - 通过 `strings.Join`，将多个地址用逗号连接成一个字符串。
	//  - 例如：`"pulsar://127.0.0.1:6650,pulsar://127.0.0.2:6650"`。
	url := strings.Join(addr, ",")
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               url,   // // Pulsar 集群的连接地址
		OperationTimeout:  30 * time.Second,  // 操作超时时间
		ConnectionTimeout: 30 * time.Second,  // 连接超时时间
		Logger:            pulsarLog.NewLoggerWithLogrus(logger),  // // 日志记录器
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

// 表示一个 Pulsar 消息队列中的主题（Topic）
// 并将主题的相关信息（租户、命名空间、主题名称）封装在一起，便于管理和操作
type Topic struct {
	Tenant    string   // 租户
	Namespace string。 // 命名空间
	Topic     string   // 主题名称
}

func (t Topic) BuildTopic() string {
	return "persistent://" + t.Tenant + "/" + t.Namespace + "/" + t.Topic
}

// 使用const定义常量
// 使用括号 `()` 将多个常量分组在一起
const (
	PublicTenant          = "public"
	GexNamespace          = "trade"
	MatchSourceTopic      = "match_source"
	MatchResultTopic      = "match_result"
	MatchResultAccountSub = "MatchResultAccountSub"
	MatchSourceSub        = "match_source_sub"
	MatchResultOrderSub   = "MatchResultOrderSub"
	MatchResultKlineSub   = "MatchResultKlineSub"
	MatchResultTickerSub  = "MatchResultTickerSub"
	MatchResultMatchSub   = "MatchResultMatchSub"
)
