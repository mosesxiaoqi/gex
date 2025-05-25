// 定义了当前文件属于 `config` 包。
//  - 这个包通常用于存放服务的配置相关代码。
//  - 在其他地方可以通过 `import "path/to/config"` 引入这个包
package config

import (
	"github.com/luxun9527/gex/common/pkg/etcd"
	logger "github.com/luxun9527/zlog"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	LoggerConfig     logger.Config
	AccountRpcConf   zrpc.RpcClientConf
	LanguageEtcdConf etcd.EtcdConfig
}
