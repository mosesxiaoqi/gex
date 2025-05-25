package etcd

import (
	"context"  // Go 标准库，用于 上下文传递（Context）
	/*一个非常实用的类型转换库，来自 spf13（Go 生态重要人物）。
	作用：安全地将任意类型转换成 string、int、bool 等常见类型。*/
	"github.com/spf13/cast"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/netx"  // netx 是 go-zero 框架的网络工具库，提供一些网络辅助方法
	/* etcd 官方提供的 v3 版本客户端库，用于：
	•	连接 etcd
	•	设置/获取/监听键值
	•	服务注册与发现 */
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"  // etcd v3 的 服务发现工具包，用于将服务注册到 etcd，并被客户端发现
	/*来自 gRPC 官方包，用于在 gRPC 中给连接或节点添加 额外属性（键值对）。
	这些属性可以在负载均衡器中用来决定如何选择连接 */
	"google.golang.org/grpc/attributes"
)

type EtcdRegisterConf struct {
	EtcdConf EtcdConfig  // etcd客户端的连接配置
	Key      string      // 注册到etcd服务key, 一般用于服务发现
	Value    string                 `json:",optional"`  // 服务地址信息 如127.1.1.1:8080
	Port     int32                  `json:",optional"`  //服务端口号，如果value为配置，可通过ip+port拼接地址
	MataData *attributes.Attributes `json:",optional"`  // grpc连接到元数据，常用于负载均衡的判断
}

func Register(conf EtcdRegisterConf) {
	go func() {
		cli, err := conf.EtcdConf.NewEtcdClient()
		if err != nil {
			logx.Severef("etcd new client err: %v", err)
		}
		manager, err := endpoints.NewManager(cli, conf.Key)
		if err != nil {
			logx.Severef("etcd new manager err: %v", err)
		}
		//设置租约时间
		resp, err := cli.Grant(context.Background(), 30)
		if err != nil {
			logx.Severef("etcd grant err: %v", err)
		}
		if conf.Value == "" {
			conf.Value = netx.InternalIp() + ":" + cast.ToString(conf.Port)
		}
		if err := manager.AddEndpoint(context.Background(), conf.Key+"/"+cast.ToString(int64(resp.ID)), endpoints.Endpoint{Addr: conf.Value, Metadata: conf.MataData}, clientv3.WithLease(resp.ID)); err != nil {
			logx.Severef("etcd add endpoint err: %v", err)
		}
		c, err := cli.KeepAlive(context.Background(), resp.ID)
		if err != nil {
			logx.Severef("etcd keepalive err: %v", err)
		}
		logx.Infof("etcd register success,key: %v,value: %v", conf.Key, conf.Value)
		for {
			select {
			case _, ok := <-c:
				if !ok {
					logx.Errorf("etcd keepalive failed,please check etcd key %v existed", conf.Key)
					return
				}
			}
		}

	}()

}
