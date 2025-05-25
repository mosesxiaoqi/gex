package consumer

import (
	"context"
	"github.com/luxun9527/gex/app/match/mq/internal/dao/model"
	"github.com/luxun9527/gex/app/match/mq/internal/logic"
	"github.com/luxun9527/gex/app/match/mq/internal/svc"
	matchMq "github.com/luxun9527/gex/common/proto/mq/match"
	"github.com/luxun9527/gex/common/utils"
	logger "github.com/luxun9527/zlog"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

func InitConsumer(sc *svc.ServiceContext) {
	go func() {
		for {
			// sc.MatchConsumer 是一个 pulsar.Consumer 实例
			// Receive(ctx) 是 Pulsar Go SDK 中的方法，用来阻塞式地接收消息，直到收到或超时
			message, err := sc.MatchConsumer.Receive(context.Background())
			if err != nil {
				logx.Errorw("consumer message match result failed", logger.ErrorField(err))
				continue
			}
			//  m 是用来接收一条撮合响应消息的变量
			var m matchMq.MatchResp
			/*if err := ...; err != nil等同于-》 err := proto.Unmarshal(...)
												if err != nil {
																...
												} */
			// 使用 Protobuf 的反序列化函数
			// proto.Unmarshal：来自 google.golang.org/protobuf/proto，是将 Protobuf 数据 → Go 对象
			// message.Payload() 这个字节数组反序列化成 m 这个结构体
			if err := proto.Unmarshal(message.Payload(), &m); err != nil {
				logx.Errorw("unmarshal match result failed", logger.ErrorField(err))
				// Ack(message)：告诉 Pulsar 这条消息我已经处理完了，不要再发给我了
				// 如果不调用 Ack，这条消息可能会：
				//	•	在超时后被认为消费失败；
				//	•	被重新投递；
				//	•	或进入死信队列。
				if err := sc.MatchConsumer.Ack(message); err != nil {
					logx.Errorw("consumer message failed", logger.ErrorField(err))
				}
				continue
			}

			m.MessageId = "matchrpc_" + m.MessageId
			logx.Infow("receive match result message", logx.Field("message", &m))
			//重复提交校验
			existed, err := sc.RedisClient.ExistsCtx(context.Background(), m.MessageId)
			if err != nil {
				logx.Errorw("redis exists failed", logger.ErrorField(err))
				continue
			}
			if existed {
				logx.Sloww("match result message id already exists", logx.Field("message_id", m.MessageId))
				if err := sc.MatchConsumer.Ack(message); err != nil {
					logx.Errorw("ack message failed", logger.ErrorField(err))
				}
				continue
			}
			storeConsumedMessageId := func() error {
				//保存7天
				if err := sc.RedisClient.SetexCtx(context.Background(), m.MessageId, "", 86400*7); err != nil {
					logx.Errorw("redis setex failed", logger.ErrorField(err))
					return err
				}
				return nil
			}
			switch r := m.Resp.(type) {
			case *matchMq.MatchResp_MatchResult:
				logx.Debugw("receive match result data ", logx.Field("data", r))
				if err := logic.NewStoreMatchResultLogic(sc).StoreMatchResult(r, storeConsumedMessageId); err != nil {
					logx.Severef("consumer match result failed err = %v", err)
				}

				matchData := &model.MatchData{
					MessageID:  message.ID(),
					MatchTime:  r.MatchResult.MatchTime,
					Volume:     utils.NewFromStringMaxPrec(r.MatchResult.Amount).Mul(utils.NewFromStringMaxPrec("2")),
					Amount:     utils.NewFromStringMaxPrec(r.MatchResult.Qty).Mul(utils.NewFromStringMaxPrec("2")),
					StartPrice: utils.NewFromStringMaxPrec(r.MatchResult.BeginPrice),
					EndPrice:   utils.NewFromStringMaxPrec(r.MatchResult.EndPrice),
					Low:        utils.NewFromStringMaxPrec(r.MatchResult.LowPrice),
					High:       utils.NewFromStringMaxPrec(r.MatchResult.HighPrice),
				}
				sc.MatchDataChan <- matchData
			}

			if err := sc.MatchConsumer.Ack(message); err != nil {
				logx.Errorw("ack message failed", logger.ErrorField(err))
			}
		}
	}()

}
