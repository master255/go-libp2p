package internal

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/p2p/transport/webrtc/internal/encoding"
	pb "github.com/libp2p/go-libp2p/p2p/transport/webrtc/pb"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multibase"
	mh "github.com/multiformats/go-multihash"
	"github.com/pion/datachannel"
	"github.com/pion/webrtc/v3"
	"google.golang.org/protobuf/proto"
)

func DecodeRemoteFingerprint(maddr ma.Multiaddr) (*mh.DecodedMultihash, error) {
	remoteFingerprintMultibase, err := maddr.ValueForProtocol(ma.P_CERTHASH)
	if err != nil {
		return nil, err
	}
	_, data, err := multibase.Decode(remoteFingerprintMultibase)
	if err != nil {
		return nil, err
	}
	return mh.Decode(data)
}

func EncodeDTLSFingerprint(fp webrtc.DTLSFingerprint) (string, error) {
	digest, err := encoding.DecodeInterpersedHexFromASCIIString(fp.Value)
	if err != nil {
		return "", err
	}
	encoded, err := mh.Encode(digest, mh.SHA2_256)
	if err != nil {
		return "", err
	}
	return multibase.Encode(multibase.Base64url, encoded)
}

// only use this if the datachannels are detached, since the OnOpen callback
// will be called immediately. Only use after the peerconnection is open.
// The context should close if the peerconnection underlying the datachannel
// is closed.
func GetDetachedChannel(ctx context.Context, dc *webrtc.DataChannel) (rwc datachannel.ReadWriteCloser, err error) {
	done := make(chan struct{})
	dc.OnOpen(func() {
		defer close(done)
		rwc, err = dc.Detach()
	})
	// this is safe since for detached datachannels, the peerconnection runs the onOpen
	// callback immediately if the SCTP transport is also connected.
	select {
	case <-done:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return
}

func AwaitPeerConnectionOpen(ufrag string, pc *webrtc.PeerConnection) <-chan error {
	errC := make(chan error)
	var once sync.Once
	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
			once.Do(func() { close(errC) })
		case webrtc.PeerConnectionStateFailed:
			once.Do(func() {
				// this ensures that we don't block this routine if the
				// listener goes away
				select {
				case errC <- fmt.Errorf("peerconnection failed: %s", ufrag):
					close(errC)
				default:
					log.Error("could not signal peerconnection failure")
				}
			})
		case webrtc.PeerConnectionStateDisconnected:
			// the connection can move to a disconnected state and back to a connected state without ICE renegotiation.
			// This could happen when underlying UDP packets are lost, and therefore the connection moves to the disconnected state.
			// If the connection then receives packets on the connection, it can move back to the connected state.
			// If no packets are received until the failed timeout is triggered, the connection moves to the failed state.
			log.Warn("peerconnection disconnected")
		}
	})
	return errC
}

// writeMessage writes a length-prefixed protobuf message to the datachannel. It
// is preferred over protoio DelimitedWriter because it is thread safe, and the
// buffer is only allocated from the global pool when writing.
func WriteMessage(rwc datachannel.ReadWriteCloser, msg *pb.Message) (int, error) {
	buf := make([]byte, 5)
	varintLen := binary.PutUvarint(buf, uint64(proto.Size(msg)))
	buf = buf[:varintLen]
	_, err := proto.MarshalOptions{}.MarshalAppend(buf, msg)
	if err != nil {
		return 0, err
	}
	_, err = rwc.Write(buf)
	if err != nil {
		return 0, err
	}
	return len(msg.Message), nil
}