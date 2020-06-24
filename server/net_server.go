package server

import (
	"bufio"
	"context"
	"io"
	"net"
	"strings"
	"time"

	"github.com/ipiao/meim/conf"
	"github.com/ipiao/meim/libs/bytes"
	xtime "github.com/ipiao/meim/libs/time"
	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
)

const (
	maxInt = 1<<31 - 1
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
		conn net.Conn
		err  error
		r    int
	)
	for {
		if conn, err = lis.Accept(); err != nil {
			// if listener close then return
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

		go serveConn(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

// 处理连接
func serveConn(s *Server, conn net.Conn, r int) {
	var (
		// timer
		tr = s.round.Timer(r)
		rp = s.round.Reader(r)
		wp = s.round.Writer(r)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	log.Infof("start serve \"%s\" with \"%s\"", lAddr, rAddr)
	s.ServeConn(conn, rp, wp, tr)
}

// ServeConn 具体处理连接中的消息
func (s *Server) ServeConn(conn net.Conn, rp, wp *bytes.Pool, tr *xtime.Timer) {
	var (
		err     error
		rid     string
		accepts []int32
		hb      time.Duration
		p       *protocol.Proto
		b       *Bucket
		trd     *xtime.TimerData
		lastHb  = time.Now()
		rb      = rp.Get()
		wb      = wp.Get()
		ch      = NewChannel(*s.c.Channel)
		rr      = &ch.Reader
		wr      = &ch.Writer
	)
	ch.Reader.ResetBuffer(conn, rb.Bytes())
	ch.Writer.ResetBuffer(conn, wb.Bytes())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// handshake
	step := 0
	authed := false
	trd = tr.Add(time.Duration(s.c.Protocol.HandshakeTimeout), func() {
		conn.Close()
		log.Errorf("key: %s remoteIP: %s step: %d tcp handshake timeout", ch.Key, conn.RemoteAddr().String(), step)
	})

	ch.IP, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	// must not setadv, only used in auth
	step = 1
	if p, err = ch.CliProto.Set(); err == nil {
		for {
			if p, err = protocol.ReadFrom(rr); err != nil {
				return
			}
			if p.Op == protocol.OpAuth {
				break
			} else {
				log.Warnf("tcp request operation(%d) not auth", p.Op)
			}
		}
		if authed = Handler.AuthChannel(ch, p); authed {
			//b=
		}

		if ch.Mid, ch.Key, rid, accepts, hb, err = s.authTCP(ctx, rr, wr, p); err == nil {
			ch.Watch(accepts...)
			b = s.Bucket(ch.Key)
			err = b.Put(rid, ch)
			if conf.Conf.Debug {
				log.Infof("tcp connected key:%s mid:%d proto:%+v", ch.Key, ch.Mid, p)
			}
		}
	}
	step = 2
	if err != nil {
		conn.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		log.Errorf("key: %s handshake failed error(%v)", ch.Key, err)
		return
	}
	trd.Key = ch.Key
	tr.Set(trd, hb)

	step = 3
	// handshake ok start dispatch goroutine
	go s.dispatchConn(conn, wr, wp, wb, ch)
	serverHeartbeat := s.RandServerHearbeat()
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		if p, err = protocol.ReadFrom(rr); err != nil {
			break
		}
		if p.Op == grpc.OpHeartbeat {
			tr.Set(trd, hb)
			p.Op = grpc.OpHeartbeatReply
			p.Body = nil
			// NOTE: send server heartbeat for a long time
			if now := time.Now(); now.Sub(lastHb) > serverHeartbeat {
				if err1 := s.Heartbeat(ctx, ch.Mid, ch.Key); err1 == nil {
					lastHb = now
				}
			}
			if conf.Conf.Debug {
				log.Infof("tcp heartbeat receive key:%s, mid:%d", ch.Key, ch.Mid)
			}
			step++
		} else {
			if err = s.Operate(ctx, p, ch, b); err != nil {
				break
			}
		}
		ch.CliProto.SetAdv()
		ch.Signal()
	}
	if err != nil && err != io.EOF && !strings.Contains(err.Error(), "closed") {
		log.Errorf("key: %s server tcp failed error(%v)", ch.Key, err)
	}
	b.Del(ch)
	tr.Del(trd)
	rp.Put(rb)
	conn.Close()
	ch.Close()
	if err = s.Disconnect(ctx, ch.Mid, ch.Key); err != nil {
		log.Errorf("key: %s mid: %d operator do disconnect error(%v)", ch.Key, ch.Mid, err)
	}
	if conf.Conf.Debug {
		log.Infof("tcp disconnected key: %s mid: %d", ch.Key, ch.Mid)
	}
}

// dispatchConn 负责将消息写会到conn中，是write线程
func (s *Server) dispatchConn(conn net.Conn, wr *bufio.Writer, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	var (
		err    error
		finish bool
	)
	if conf.Conf.Debug {
		log.Infof("key: %s start dispatch tcp goroutine", ch.Key)
	}
	for {
		var p = ch.Ready()
		if conf.Conf.Debug {
			log.Infof("key:%s dispatch msg:%v", ch.Key, *p)
		}
		switch p {
		case SignalFinish:
			if conf.Conf.Debug {
				log.Infof("key: %s wakeup exit dispatch goroutine", ch.Key)
			}
			finish = true
			goto failed
		case SignalReady:
			// fetch message from svrbox(client send)
			for {
				if p, err = ch.CliProto.Get(); err != nil {
					break
				}
				// TODO
				p.Body = nil // avoid memory leak
				ch.CliProto.GetAdv()
			}
		default:
			// server send
			if err = protocol.WriteTo(wr, p); err != nil {
				goto failed
			}
			if conf.Conf.Debug {
				log.Infof("tcp sent a message key:%s mid:%d proto:%+v", ch.Key, ch.Mid, p)
			}
		}
		// only hungry flush response
		if err = wr.Flush(); err != nil {
			break
		}
	}
failed:
	if err != nil {
		log.Errorf("key: %s dispatch tcp error(%v)", ch.Key, err)
	}

	conn.Close()
	wp.Put(wb)
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = ch.Ready() == SignalFinish
	}
	if conf.Conf.Debug {
		log.Infof("key: %s dispatch goroutine exit", ch.Key)
	}
}
