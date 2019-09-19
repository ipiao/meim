package registrymq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
	"github.com/ipiao/meim/util"
	"github.com/streadway/amqp"
)

type Message = protocol.InternalMessage

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
func (s session) Close() error {
	if s.Connection == nil {
		return nil
	}
	return s.Connection.Close()
}

// rabbit配置
type RabbitBrokenConfig struct {
	Node         string
	ExchangeName string
	ExchangeKind string
	RPCTimeout   time.Duration
	Url          string
	Chansize     int
	Channels     uint64
	QueuePrefix  string // 队列前缀
}

// 用户rpc操作的请求结构
type request struct {
	node string
	msg  *Message
	ret  chan []byte
}

// type RPCResponse struct {
// 	Code int32
// 	Msg  string
// 	Data interface{}
// }

// 一个rabbit完整的broken
type RabbitBroken struct {
	cancel         context.CancelFunc
	cfg            *RabbitBrokenConfig
	subMessageChan chan *Message // sub message
	pubMessageChan chan *request // pub message
	rpcRequestChan chan *request // rpc message
	bp             *util.BufferPool
	rpcHandler     func(*Message) []byte
	DC             protocol.DataCreator
}

const (
	ChannelPub = 1 << iota
	ChannelSub
	ChannelRPC
	ChannelRPCServer
	AllChannel = ChannelPub | ChannelSub | ChannelRPC | ChannelRPCServer
)

// 新建rabbot通道
func NewRabbitBroken(cfg *RabbitBrokenConfig, rpcHandler func(*Message) []byte) *RabbitBroken {
	if cfg.QueuePrefix == "" {
		cfg.QueuePrefix = "router_message"
	}
	ctx, done := context.WithCancel(context.Background())
	rb := &RabbitBroken{
		cancel:     done,
		cfg:        cfg,
		bp:         util.NewBufferPool(),
		rpcHandler: rpcHandler,
	}

	if cfg.Channels&ChannelSub != 0 {
		go func() {
			rb.subMessageChan = make(chan *Message, cfg.Chansize)
			rb.subscribe(redial(ctx, cfg.Url, rb.cfg.ExchangeName, rb.cfg.ExchangeKind))
			done()
		}()
	}

	if cfg.Channels&ChannelPub != 0 {
		go func() {
			rb.pubMessageChan = make(chan *request, cfg.Chansize)
			rb.publish2(redial(ctx, cfg.Url, rb.cfg.ExchangeName, rb.cfg.ExchangeKind))
			done()
		}()
	}

	if cfg.Channels&ChannelRPCServer != 0 {
		if rb.rpcHandler == nil {
			log.Fatalf("[rabbit] listening RPCServer, rpcHandler is not set")
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

// 生成队列名
func (rb *RabbitBroken) getQueueName() string {
	return fmt.Sprintf("%s_%s", rb.cfg.QueuePrefix, rb.cfg.Node)
}

// 队列绑定key
func (rb *RabbitBroken) getBindKey() string {
	return fmt.Sprintf("%s.%s.*", rb.cfg.QueuePrefix, rb.cfg.Node)
}

// rpc队列名
func (rb *RabbitBroken) getRpcQueueName() string {
	return fmt.Sprintf("%s_rpc_%s", rb.cfg.QueuePrefix, rb.cfg.Node)
}

// 队列绑定key
func (rb *RabbitBroken) getRpcBindKey() string {
	return fmt.Sprintf("%s_rpc.%s.*", rb.cfg.QueuePrefix, rb.cfg.Node)
}

// 路由key
func (rb *RabbitBroken) getRoutingKey(node string, message *Message) string {
	return fmt.Sprintf("%s.%s.%d", rb.cfg.QueuePrefix, node, message.Header.Cmd())
}

// 路由key
func (rb *RabbitBroken) getRpcRoutingKey(node string, message *Message) string {
	return fmt.Sprintf("%s_rpc.%s.%d", rb.cfg.QueuePrefix, node, message.Header.Cmd())
}

// 发送消息
func (rb *RabbitBroken) publish(sessions chan chan session) {
	for session := range sessions {
		pub := <-session
		if !pub.connected() {
			log.Warnf("[rabbit] session not connected")
			time.Sleep(time.Millisecond * 10)
			continue
		}

		var (
			running bool
			req     *request
			reading = rb.pubMessageChan
			pending = make(chan *request, 1)
			confirm = make(chan amqp.Confirmation, 1)
		)

		if err := pub.Confirm(false); err != nil {
			log.Errorf("[rabbit] publisher confirms not supported")
			close(confirm)
		} else {
			pub.NotifyPublish(confirm)
		}

		log.Infof("[rabbit] publishing...")

	Publish:
		for {
			var body []byte
			select {
			case confirmed, ok := <-confirm:
				if !ok {
					break Publish
				}
				if !confirmed.Ack {
					log.Debugf("[rabbit] nack message %d, body: %q", confirmed.DeliveryTag, string(body))
				}
				reading = rb.pubMessageChan

			case req = <-pending:
				body := rb.encodeMessage(req.msg)
				routineKey := rb.getRoutingKey(req.node, req.msg)
				err := pub.Publish(rb.cfg.ExchangeName, routineKey, false, false, amqp.Publishing{
					Body: body,
				})
				// Retry failed delivery on the next session
				if err != nil {
					log.Errorf("[rabbit] can not publish message: %v", err)
					pending <- req
					pub.Close()
					break Publish
				}

			case req, running = <-reading:
				if !running {
					return
				}
				pending <- req
				reading = nil
			}
		}
	}
}

// 发送消息
func (rb *RabbitBroken) publish2(sessions chan chan session) {
	for session := range sessions {
		pub := <-session
		if !pub.connected() {
			log.Warnf("[rabbit] session not connected")
			time.Sleep(time.Millisecond * 10)
			continue
		}
		var (
			reading = rb.pubMessageChan
			body    []byte
			err     error
		)
		log.Infof("[rabbit] publishing...")

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
					pub.Close()
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
		if _, err := sub.QueueDeclare(queue, false, false, false, false, nil); err != nil {
			log.Errorf("[rabbit] cannot consume from exclusive queue: %q, %v", queue, err)
			sub.Close()
			continue
		}

		if err := sub.QueueBind(queue, rb.getBindKey(), rb.cfg.ExchangeName, false, nil); err != nil {
			log.Errorf("[rabbit] cannot consume without a binding to exchange: %q, %v", rb.cfg.ExchangeName, err)
			sub.Close()
			continue
		}

		deliveries, err := sub.Consume(queue, "", true, true, false, false, nil)
		if err != nil {
			log.Errorf("[rabbit]  cannot consume from: %q, %v", queue, err)
			sub.Close()
			continue
		}

		log.Infof("[rabbit] subscribed...")

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

		if _, err := rpc.QueueDeclare(rpcQueueName, false, false, false, false, nil); err != nil {
			log.Errorf("[rabbit] cannot consume from exclusive queue: %q, %v", rpcQueueName, err)
			rpc.Close()
			continue
		}

		if err := rpc.QueueBind(rpcQueueName, rb.getRpcBindKey(), rb.cfg.ExchangeName, false, nil); err != nil {
			log.Errorf("[rabbit] cannot consume without a binding to exchange: %q, %v", rb.cfg.ExchangeName, err)
			rpc.Close()
			continue
		}

		msgs, err := rpc.Consume(rpcQueueName, "", true, false, false, false, nil)
		if err != nil {
			log.Errorf("[rabbit] cannot consume from exclusive queue: %q, %v", rpcQueueName, err)
			rpc.Close()
			continue
		}

		for d := range msgs {
			message := rb.decodeMessage(d.Body)
			body := rb.rpcHandler(message)

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
					rpc.Close()
					reqs <- req
					break PUBLISH
				}

				if err := rpc.QueueBind(q.Name, q.Name, rb.cfg.ExchangeName, false, nil); err != nil {
					log.Errorf("[rabbit] cannot consume without a binding to exchange: %q, %v", rb.cfg.ExchangeName, err)
					rpc.Close()
					reqs <- req
					break PUBLISH
				}

				msgs, err := rpc.Consume(q.Name, "", true, true, false, false, nil)
				if err != nil {
					log.Errorf("[rabbit] cannot consume from: %q, %v", rb.getRpcQueueName(), err)
					rpc.Close()
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
					rpc.Close()
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

func (rb *RabbitBroken) encodeMessage(msg *Message) []byte {
	buf := rb.bp.Get()
	defer rb.bp.Put(buf)

	protocol.WriteInternalMessage(buf, msg)
	return buf.Bytes()
}

func (rb *RabbitBroken) decodeMessage(body []byte) *Message {
	buf := rb.bp.Get()
	defer rb.bp.Put(buf)

	buf.Write(body)
	message, _ := protocol.ReadInternalMessage(buf, rb.DC)
	return message
}

func (rb *RabbitBroken) Node() string {
	return rb.cfg.Node
}

// 异步发送消息
// 异步发送消息
func (rb *RabbitBroken) SendMessage(node string, msg *Message) error {
	if rb.pubMessageChan == nil {
		return errors.New("not registered")
	}
	if msg == nil {
		return errors.New("nil message")
	}
	log.Debugf("[rabbit] SendMessage-> %s : %v", node, msg)
	rb.pubMessageChan <- &request{
		node: node,
		msg:  msg,
	}
	return nil
}

func (rb *RabbitBroken) SendMessageTimeout(node string, msg *Message, timeout time.Duration) error {
	if rb.pubMessageChan == nil {
		return errors.New("not registered")
	}
	if msg == nil {
		return errors.New("nil message")
	}
	log.Debugf("[rabbit] SendMessageTimeout-> %s : %v", node, msg)
	req := &request{
		node: node,
		msg:  msg,
	}
	var err error
	select {
	case rb.pubMessageChan <- req:
	case <-time.After(timeout):
		log.Infof("[rabbit] send message timeout,msg %v dropped", msg)
		err = errors.New("send message timeout")
	}
	return err
}

// rpc 服务调用
// 同步发送等待返回
func (rb *RabbitBroken) SyncMessage(node string, msg *Message) ([]byte, error) {
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
		return b, nil
	case <-time.After(rb.cfg.RPCTimeout):
		return nil, errors.New("syncmessage timeout")
	}
}

// 返回接收通道
func (rb *RabbitBroken) ReceiveMessage() *Message {
	return <-rb.subMessageChan
}

// 关闭
func (rb *RabbitBroken) Close() error {
	if rb.cancel != nil {
		rb.cancel()
	}
	return nil
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
