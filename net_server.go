package meim

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/ipiao/meim/conf"

	xtime "github.com/ipiao/meim/libs/time"
	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
)

// InitNetworks 开启监听连接
func InitNetListeners(server *Server) (err error) {
	var (
		listener net.Listener
	)
	for _, network := range server.c.Networks {
		creator := LoadListenerCreator(network.Protocol)
		if creator == nil {
			log.Errorf("protocol listener creator not registered: %s", network.Protocol)
			continue
		}
		for _, bind := range network.Bind {
			for i := 0; i < network.Accept; i++ {
				listener, err = creator(network, bind)
				if err != nil {
					log.Errorf("create protocol listener failed: %v", err)
					return
				}
				go acceptNetListener(server, network, listener)
			}
		}
	}
	return
}

// acceptNetListener 接收连接并进行事先处理
func acceptNetListener(server *Server, network *conf.Network, lis net.Listener) {
	var (
		conn    net.Conn
		err     error
		retries int
	)
	for {

		select {
		case <-server.closed:
			_ = lis.Close()
			return
		default:
		}

		if conn, err = lis.Accept(); err != nil {
			// if listener close then return
			retries = 0
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				delay := server.backoff.BackOff(retries)
				retries++
				log.Warnf("listener.Accept(\"%s\") error(%v)\nretrying in %s", lis.Addr().String(), err, delay)
				time.Sleep(delay)
				continue
			}
			log.Errorf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if err = tcpConn.SetKeepAlive(network.KeepAlive); err != nil {
				log.Errorf("conn.SetKeepAlive() error(%v)", err)
				return
			}
			if err = tcpConn.SetReadBuffer(network.Rcvbuf); err != nil {
				log.Errorf("conn.SetReadBuffer() error(%v)", err)
				return
			}
			if err = tcpConn.SetWriteBuffer(network.Sndbuf); err != nil {
				log.Errorf("conn.SetWriteBuffer() error(%v)", err)
				return
			}
		}

		go serveConn(server, conn)
	}
}

func serveConn(s *Server, conn net.Conn) {
	var (
		cid   = int(s.unid.Generate())
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
		rp    = s.round.Reader(cid)
		wp    = s.round.Writer(cid)
	)
	log.Infof("start serve \"%s\" with \"%s\", cid %d", lAddr, rAddr, cid)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := NewChannel(ctx, conn, rp.Get(), wp.Get(), cid, s.c.Channel.SvrProto, s.c.Channel.CliProto)
	_ = s.ServeChannel(ch)
}

// ServeConn 具体处理连接中的消息
func (s *Server) ServeChannel(ch *conf.Channel) (err error) {
	// 首先握手，握手失败了直接返回
	// 握手成功了进入正常流程
	if err = s.handshakeChannel(ch); err != nil {
		return
	}
	s.wg.Add(1)
	// 握手成功后的执行流程
	var (
		// timer
		p *protocol.Proto
		//lastHb  = time.Now()
		rr = &ch.Reader
	)

	// 异步处理写
	go s.dispatchChannel(ch)
	// 处理读
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			log.Warnf("channel ring set failed: %s", ch)
			break
		}
		if p, err = protocol.ReadFrom(rr); err != nil {
			// 处理非连接错误
			if _, ok := err.(net.Error); ok || err == io.EOF {
				log.Debugf("channel read proto failed: %s", ch)
			} else {
				log.Errorf("channel read proto failed: %s", ch)
			}
			log.Warnf("channel read proto failed: %s,\n %s", ch, p)
			break
		}
		if err = Handler.HandleProto(ch, p); err != nil {
			log.Warnf("channel handle proto failed: %s,\n %s", ch, p)
			break
		}
		ch.CliProto.SetAdv()
		ch.Signal()
	}

	s.round.Reader(ch.CID).Put(ch.rb) // 读线程显示归还读资源
	ch.Close()                        // 给写通知写线程结束
	return
}

// dispatchConn 负责将消息写会到conn中，是write线程
func (s *Server) dispatchChannel(ch *conf.Channel) {
	var (
		err    error
		finish bool
		wr     = &ch.Writer
	)
	log.Debugf("channel: %s, start dispatch tcp goroutine", ch.String())
	for {
		var p = ch.Ready()
		switch p {
		case SignalFinish:
			log.Debugf("channel: %s, wakeup exit dispatch goroutine", ch)
			finish = true
			goto failed
		case SignalReady:
			// 读到请求返回，从proto中获取返回
			for {
				if p, err = ch.CliProto.Get(); err != nil {
					break
				}
				log.Debugf("channel %s write back response %s", ch, p)
				if err = protocol.WriteTo(wr, p); err != nil {
					goto failed
				}
				p.Body = nil // avoid memory leak
				ch.CliProto.GetAdv()
			}
		default:
			// server send
			log.Debugf("channel: %s, sent a message %s", ch, p)
			if err = protocol.WriteTo(wr, p); err != nil {
				goto failed
			}
		}
		// 直接刷缓存
		if err = wr.Flush(); err != nil {
			break
		}
	}
failed:
	if err != nil {
		log.Errorf("channel: %s dispatch tcp error(%v)", ch, err)
	}

	// 在实际结束前，处理一些任务
	Handler.HandleClosed(ch)

	// 一般是客户端主动断开，所以是先读退出，然后处理写退出，所以在写线程做最后的关闭处理
	s.Bucket(ch.Key).Del(ch) //
	_ = ch.conn.Close()      // 关闭底层连接，这样读线程也会结束

	s.round.Writer(ch.CID).Put(ch.wb) // 归还写资源
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = ch.Ready() == SignalFinish
	}
	s.wg.Done()
	log.Infof("key: %s dispatch goroutine exit", ch.Key)
}

func (s *Server) handshakeChannel(ch *conf.Channel) (err error) {
	// ip addr
	var (
		trd *xtime.TimerData
		// timer
		p  *protocol.Proto
		tr = s.round.Timer(ch.CID)
		rp = s.round.Reader(ch.CID)
		wp = s.round.Writer(ch.CID)
		rr = &ch.Reader
	)

	// handshake
	step := 0
	trd = tr.Add(time.Duration(s.c.Protocol.HandshakeTimeout), func() {
		_ = ch.conn.Close()
		log.Errorf("channel %s handshake timeout: %d", ch, step)
	})

	// must not setadv, only used in auth
	// 握手认证过程
	step = 1
	if p, err = ch.CliProto.Set(); err == nil {
		for {
			if p, err = protocol.ReadFrom(rr); err != nil {
				return
			}
			if p.Op == protocol.OpAuth {
				break
			} else {
				log.Warnf("channel %s request operation(%d) not auth", ch, p.Op)
			}
		}
		var rid string
		var accepts []int32
		if ch.Mid, ch.Key, rid, accepts, _, err = Handler.HandleAuth(ch, p); err == nil {
			ch.Watch(accepts...)
			err = s.Bucket(ch.Key).Put(rid, ch)
		}
	}

	step = 2
	tr.Del(trd)

	if err != nil {
		_ = ch.conn.Close()
		rp.Put(ch.rb)
		wp.Put(ch.wb)
		log.Warnf("channel %s, handshake failed error(%v)", ch, err)

		//Handler.HandleClosed(ch)
		//log.Infof("tcp disconnected key: %s mid: %d", ch.Key, ch.Mid)
	}
	return
}
