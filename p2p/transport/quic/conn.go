package libp2pquic

import (
	"fmt"
	"net"

	smux "github.com/jbenet/go-stream-muxer"
	tpt "github.com/libp2p/go-libp2p-transport"
	quic "github.com/lucas-clemente/quic-go"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

type quicConn struct {
	sess      quic.Session
	transport tpt.Transport

	laddr ma.Multiaddr
	raddr ma.Multiaddr
}

var _ tpt.Conn = &quicConn{}
var _ tpt.MultiStreamConn = &quicConn{}

func newQuicConn(sess quic.Session, t tpt.Transport) (*quicConn, error) {
	// analogues to manet.WrapNetConn

	laddr, err := quicMultiAddress(sess.LocalAddr())
	if err != nil {
		return nil, fmt.Errorf("failed to convert nconn.LocalAddr: %s", err)
	}

	// analogues to manet.WrapNetConn
	raddr, err := quicMultiAddress(sess.RemoteAddr())
	if err != nil {
		return nil, fmt.Errorf("failed to convert nconn.RemoteAddr: %s", err)
	}

	return &quicConn{
		sess:      sess,
		laddr:     laddr,
		raddr:     raddr,
		transport: t,
	}, nil
}

func (c *quicConn) AcceptStream() (smux.Stream, error) {
	str, err := c.sess.AcceptStream()
	if err != nil {
		return nil, err
	}
	return &quicStream{Stream: str}, nil
}

func (c *quicConn) OpenStream() (smux.Stream, error) {
	str, err := c.sess.OpenStream()
	if err != nil {
		return nil, err
	}
	return &quicStream{Stream: str}, nil
}

func (c *quicConn) Serve(handler smux.StreamHandler) {
	for { // accept loop
		s, err := c.AcceptStream()
		if err != nil {
			return // err always means closed.
		}
		go handler(s)
	}
}

func (c *quicConn) Close() error {
	return c.sess.Close(nil)
}

// TODO: implement this
func (c *quicConn) IsClosed() bool {
	return false
}

func (c *quicConn) LocalAddr() net.Addr {
	return c.sess.LocalAddr()
}

func (c *quicConn) LocalMultiaddr() ma.Multiaddr {
	return c.laddr
}

func (c *quicConn) RemoteAddr() net.Addr {
	return c.sess.RemoteAddr()
}

func (c *quicConn) RemoteMultiaddr() ma.Multiaddr {
	return c.raddr
}

func (c *quicConn) Transport() tpt.Transport {
	return c.transport
}

// TODO: there must be a better way to do this
func quicMultiAddress(na net.Addr) (ma.Multiaddr, error) {
	udpMA, err := manet.FromNetAddr(na)
	if err != nil {
		return nil, err
	}
	quicMA, err := ma.NewMultiaddr(udpMA.String() + "/quic")
	if err != nil {
		return nil, err
	}
	return quicMA, nil
}
