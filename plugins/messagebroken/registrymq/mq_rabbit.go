package registrymq

import (
	"context"
	"errors"
	"fmt"
	"time"
	"xinxin/1v1_micro/libs/util"

	"github.com/google/uuid"
	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
	"github.com/streadway/amqp"
)

type RPCHandler func(*protocol.InternalMessage) *protocol.InternalMessage

const (
	ChannelPub = 1 << iota
	ChannelSub
	ChannelRPC
	ChannelRPCServer
	AllChannel = ChannelPub | ChannelSub | ChannelRPC | ChannelRPCServer
)

// rabbit配置
type RabbitBrokenConfig struct {
	ExchangeName string
	ExchangeKind string
	Url          string
	RPCTimeout   time.Duration
	SendTimeout  time.Duration
	Chansize     int
	Channels     uint64
	QueuePrefix  string // 队列前缀
	Node         int
}

func (cfg *RabbitBrokenConfig) init() {
	if cfg.ExchangeName == "" {
		cfg.ExchangeName = "meim"
	}
	if cfg.ExchangeKind == "" || !util.IsStringIn(cfg.ExchangeKind, []string{"direct", "fanout", "topic"}) {
		cfg.ExchangeKind = "topic"
	}
	if cfg.RPCTimeout == 0 {
		cfg.RPCTimeout = time.Second * 5
	}
	if cfg.Chansize == 0 {
		cfg.Chansize = 512
	}
	if cfg.QueuePrefix == "" {
		cfg.QueuePrefix = "message"
	}
	if cfg.Node <= 0 {
		log.Fatal("[rabbit] cfg node must be greater than 0")
	}
}

// 用户rpc操作的请求结构
type request struct {
	node int
	msg  *protocol.InternalMessage
	ret  chan []byte
}

// 一个rabbit完整的broken
type RabbitBroken struct {
	cancel         context.CancelFunc
	cfg            *RabbitBrokenConfig
	pubMessageChan chan *request                  // pub message
	rpcRequestChan chan *request                  // rpc message
	subMessageChan chan *protocol.InternalMessage // sub message
	rpcHandler     RPCHandler
	dc             protocol.DataCreator
}

// 新建rabbot通道,参数需要给定
func NewRabbitBroken(cfg *RabbitBrokenConfig, dc protocol.DataCreator, rpcHandler RPCHandler) *RabbitBroken {
	cfg.init()

	ctx, done := context.WithCancel(context.Background())
	rb := &RabbitBroken{
		cancel:     done,
		cfg:        cfg,
		rpcHandler: rpcHandler,
		dc:         dc,
	}

	if cfg.Channels&ChannelSub != 0 {
		if dc == nil {
			log.Fatal("sub broken must set DataCreator")
		}
		go func() {
			rb.subMessageChan = make(chan *protocol.InternalMessage, cfg.Chansize)
			rb.subscribe(redial(ctx, cfg.Url, rb.cfg.ExchangeName, rb.cfg.ExchangeKind))
			done()
		}()
	}

	if cfg.Channels&ChannelPub != 0 {
		go func() {
			rb.pubMessageChan = make(chan *request, cfg.Chansize)
			rb.publish(redial(ctx, cfg.Url, rb.cfg.ExchangeName, rb.cfg.ExchangeKind))
			done()
		}()
	}

	if cfg.Channels&ChannelRPCServer != 0 {
		if dc == nil {
			log.Fatal("[rabbit] RPCServer broken must set DataCreator")
		}
		if rb.rpcHandler == nil {
			log.Fatal("[rabbit] listening RPCServer, rpcHandler is not set")
		}
		go func() {
			rb.rpcServer(redial(ctx, cfg.Url, rb.cfg.ExchangeName, rb.cfg.ExchangeKind))
			done()
		}()
	}
	if cfg.Channels&ChannelRPC != 0 {
		go func() {
			rb.rpcRequestChan = make(chan *request, cfg.Chansize)
			rb.rpc(redial(ctx, cfg.Url, rb.cfg.ExchangeName, rb.cfg.ExchangeKind))
			done()
		}()
	}

	return rb
}

// 发送消息
func (rb *RabbitBroken) publish(sessions chan chan session) {
	for session := range sessions {
		pub := <-session
		if !pub.connected() {
			log.Warnf("[rabbit] session not connected")
			time.Sleep(time.Millisecond * 100)
			continue
		}
		var (
			reading = rb.pubMessageChan
			body    []byte
			err     error
		)
		log.Debug("[rabbit] publishing...")

	Publish:
		for {
			select {
			case req, ok := <-reading:
				if !ok {
					return
				}
				body = rb.encodeMessage(req.msg)
				routineKey := rb.getRoutingKey(req.node, req.msg)
				err = pub.Publish(rb.cfg.ExchangeName, routineKey, false, false, amqp.Publishing{
					Body: body,
				})
				if err != nil {
					log.Errorf("[rabbit] can not publish message: %v", err)
					reading <- req
					pub.close()
					break Publish
				}
			}
		}
	}
}

// 订阅消息
func (rb *RabbitBroken) subscribe(sessions chan chan session) {

	queue := rb.getQueueName()
	for session := range sessions {
		sub := <-session
		if !sub.connected() {
			log.Warnf("[rabbit] session not connected")
			time.Sleep(time.Millisecond * 10)
			continue
		}

		// 去除排他性
		if _, err := sub.QueueDeclare(queue, true, false, false, false, nil); err != nil {
			log.Errorf("[rabbit] cannot consume from exclusive queue: %q, %v", queue, err)
			sub.close()
			continue
		}

		if err := sub.QueueBind(queue, rb.getBindKey(), rb.cfg.ExchangeName, false, nil); err != nil {
			log.Errorf("[rabbit] cannot consume without a binding to exchange: %q, %v", rb.cfg.ExchangeName, err)
			sub.close()
			continue
		}

		deliveries, err := sub.Consume(queue, "", true, true, false, false, nil)
		if err != nil {
			log.Errorf("[rabbit]  cannot consume from: %q, %v", queue, err)
			sub.close()
			continue
		}

		log.Debug("[rabbit] subscribed...")

		for msg := range deliveries {
			message := rb.decodeMessage(msg.Body)
			rb.subMessageChan <- message
			// sub.Ack(msg.DeliveryTag, false)
		}
	}
}

// 开启rpc服务
func (rb *RabbitBroken) rpcServer(sessions chan chan session) {
	rpcQueueName := rb.getRpcQueueName()
	for session := range sessions {
		rpc := <-session
		if !rpc.connected() {
			log.Warnf("[rabbit] session not connected")
			time.Sleep(time.Millisecond * 10)
			continue
		}

		if _, err := rpc.QueueDeclare(rpcQueueName, true, false, false, false, nil); err != nil {
			log.Errorf("[rabbit] cannot consume from exclusive queue: %q, %v", rpcQueueName, err)
			rpc.close()
			continue
		}

		if err := rpc.QueueBind(rpcQueueName, rb.getRpcBindKey(), rb.cfg.ExchangeName, false, nil); err != nil {
			log.Errorf("[rabbit] cannot consume without a binding to exchange: %q, %v", rb.cfg.ExchangeName, err)
			rpc.close()
			continue
		}

		msgs, err := rpc.Consume(rpcQueueName, "", true, false, false, false, nil)
		if err != nil {
			log.Errorf("[rabbit] cannot consume from exclusive queue: %q, %v", rpcQueueName, err)
			rpc.close()
			continue
		}

		for d := range msgs {
			message := rb.decodeMessage(d.Body)
			var body []byte

			resp := rb.rpcHandler(message)
			if resp != nil {
				body = rb.encodeMessage(resp)
			}

			err = rpc.Publish(rb.cfg.ExchangeName, d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          body,
				})
		}
	}
}

// 发送消息
func (rb *RabbitBroken) rpc(sessions chan chan session) {
	for session := range sessions {
		rpc := <-session
		var reqs = rb.rpcRequestChan
	PUBLISH:
		for {
			select {
			case req := <-reqs:
				corrId := uuid.New().String()
				// 接收rpc服务返回队列
				q, err := rpc.QueueDeclare(corrId, false, true, true, false, nil)
				if err != nil {
					log.Errorf("[rabbit] cannot consume from exclusive queue: %q, %v", q.Name, err)
					rpc.close()
					reqs <- req
					break PUBLISH
				}

				if err := rpc.QueueBind(q.Name, q.Name, rb.cfg.ExchangeName, false, nil); err != nil {
					log.Errorf("[rabbit] cannot consume without a binding to exchange: %q, %v", rb.cfg.ExchangeName, err)
					rpc.close()
					reqs <- req
					break PUBLISH
				}

				msgs, err := rpc.Consume(q.Name, "", true, true, false, false, nil)
				if err != nil {
					log.Errorf("[rabbit] cannot consume from: %q, %v", rb.getRpcQueueName(), err)
					rpc.close()
					reqs <- req
					break PUBLISH
				}

				routineKey := rb.getRpcRoutingKey(req.node, req.msg)
				err = rpc.Publish(rb.cfg.ExchangeName, routineKey, false, false,
					amqp.Publishing{
						ContentType:   "text/plain",
						CorrelationId: corrId,
						ReplyTo:       q.Name,
						Body:          rb.encodeMessage(req.msg),
					})
				if err != nil {
					log.Errorf("[rabbit] can not publish message: %v", err)
					reqs <- req
					rpc.close()
					break PUBLISH
				}
				go func() {
					for {
						select {
						case d := <-msgs:
							if corrId == d.CorrelationId {
								req.ret <- d.Body
								return
							}
						case <-time.After(rb.cfg.RPCTimeout):
							log.Warnf("[rabbit] rpc timeout : %v", req.msg.Header.Cmd())
							req.ret <- nil
							return
						}
					}
				}()
			}
		}
	}
}

// 生成队列名
func (rb *RabbitBroken) getQueueName() string {
	return fmt.Sprintf("%s_%d", rb.cfg.QueuePrefix, rb.cfg.Node)
}

// 队列绑定key
func (rb *RabbitBroken) getBindKey() string {
	return fmt.Sprintf("%s.%d.*", rb.cfg.QueuePrefix, rb.cfg.Node)
}

// rpc队列名
func (rb *RabbitBroken) getRpcQueueName() string {
	return fmt.Sprintf("%s_rpc_%d", rb.cfg.QueuePrefix, rb.cfg.Node)
}

// 队列绑定key
func (rb *RabbitBroken) getRpcBindKey() string {
	return fmt.Sprintf("%s_rpc.%d.*", rb.cfg.QueuePrefix, rb.cfg.Node)
}

// 路由key
func (rb *RabbitBroken) getRoutingKey(node int, message *protocol.InternalMessage) string {
	return fmt.Sprintf("%s.%d.%d", rb.cfg.QueuePrefix, node, message.Receiver)
}

// 路由key
func (rb *RabbitBroken) getRpcRoutingKey(node int, message *protocol.InternalMessage) string {
	return fmt.Sprintf("%s_rpc.%d.%d", rb.cfg.QueuePrefix, node, message.Receiver)
}

func (rb *RabbitBroken) encodeMessage(msg *protocol.InternalMessage) []byte {
	b, _ := protocol.MarshalInternalMessage(msg)
	return b
}

func (rb *RabbitBroken) decodeMessage(body []byte) *protocol.InternalMessage {
	message, err := protocol.UnmarshalInternalMessgae(body, rb.dc)
	if err != nil {
		return nil
	}
	return message
}

func (rb *RabbitBroken) Node() int {
	return rb.cfg.Node
}

// 异步发送消息
// 异步发送消息
func (rb *RabbitBroken) SendMessage(node int, msg *protocol.InternalMessage) error {
	if rb.pubMessageChan == nil {
		return errors.New("not registered")
	}
	if msg == nil {
		return errors.New("nil message")
	}
	log.Debugf("[rabbit] SendMessage-> %s : %v", node, msg)

	if rb.cfg.SendTimeout > 0 {
		select {
		case rb.pubMessageChan <- &request{
			node: node,
			msg:  msg,
		}:
		case <-time.After(rb.cfg.SendTimeout):
			log.Infof("[rabbit] send message timeout,msg %v dropped", msg)
			return errors.New("send message timeout")
		}
	} else {
		rb.pubMessageChan <- &request{
			node: node,
			msg:  msg,
		}
	}
	return nil
}

// rpc 服务调用
// 同步发送等待返回
func (rb *RabbitBroken) SyncMessage(node int, msg *protocol.InternalMessage) (*protocol.InternalMessage, error) {
	log.Debugf("[rabbit] SyncMessage: node->%s: %+v", node, msg)
	if msg == nil {
		return nil, errors.New("nil message")
	}
	retChan := make(chan []byte, 1)
	rb.rpcRequestChan <- &request{
		node: node,
		msg:  msg,
		ret:  retChan,
	}
	select {
	case b := <-retChan:
		return rb.decodeMessage(b), nil
	case <-time.After(rb.cfg.RPCTimeout):
		return nil, errors.New("syncmessage timeout")
	}
}

// 返回接收通道
func (rb *RabbitBroken) ReceiveMessage() *protocol.InternalMessage {
	return <-rb.subMessageChan
}

// 关闭
func (rb *RabbitBroken) Close() {
	if rb.cancel != nil {
		rb.cancel()
	}
}

// 重连
func redial(ctx context.Context, url, exchange, exchangeKind string) chan chan session {
	sessions := make(chan chan session)

	go func() {
		sess := make(chan session)
		defer close(sessions)

		for {
			select {
			case sessions <- sess:
			case <-ctx.Done():
				log.Infof("[rabbit] shutting down session factory")
				return
			}

			conn, err := amqp.Dial(url)
			if err != nil {
				log.Infof("[rabbit] cannot (re)dial: %v: %q", err, url)
			}

			ch, err := conn.Channel()
			if err != nil {
				log.Infof("[rabbit] cannot create channel: %v", err)
			}

			if err := ch.ExchangeDeclare(exchange, exchangeKind, true, true, false, false, nil); err != nil {
				log.Infof("[rabbit] cannot declare %v exchange: %v", exchangeKind, err)
			}

			select {
			case sess <- session{conn, ch}:
			case <-ctx.Done():
				log.Infof("[rabbit] shutting down new session")
				return
			}
		}
	}()
	return sessions
}

// session
type session struct {
	*amqp.Connection
	*amqp.Channel
}

// 是否是成功的连接
func (s session) connected() bool {
	return s.Connection != nil && s.Channel != nil
}

// 连同连接和通道一起关闭
func (s session) close() {
	if s.Connection != nil {
		s.Connection.Close()
	}
}
