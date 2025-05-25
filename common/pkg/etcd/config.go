package etcd

// etcd 是一个 开源的分布式键值存储系统，主要用于服务发现、配置管理和一致性协调
// etcd = 分布式数据库 + 一致性协议（Raft）
/*
服务发现流程：
	1.	服务 A 启动，向 etcd 注册自己的地址：/services/A/instance1 = 10.0.0.1:8080
	2.	服务 B 想调用服务 A，会从 etcd 获取 /services/A 下的地址列表。
	3.	服务 A 崩溃，它注册的地址有 TTL（过期时间），自动从 etcd 中删除。
*/
import (
	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type EtcdConfig struct {
	// 这个结构体用于配置 etcd 连接信息，主要是 Endpoints —— etcd 节点地址数组
	Endpoints []string
}

func (c *EtcdConfig) NewEtcdClient() (*clientv3.Client, error) {
	config := clientv3.Config{Endpoints: c.Endpoints, DialTimeout: time.Second * time.Duration(5)}
	return clientv3.New(config)
}
func (c *EtcdConfig) MustNewEtcdClient() *clientv3.Client {
	client, err := c.NewEtcdClient()
	if err != nil {
		logx.Severef("etcd client init failed, err: %v", err)
	}
	return client
}
