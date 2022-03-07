package rpc

import (
	"context"
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/network/rpc/transport"
	"sync"
)

type ResponseStreams struct {
	closed      bool
	streams     map[int]*ResponseStream
	streamsLock sync.Mutex

	raw    transport.RawConnection
	logger logging.Logger
}

func NewResponseStreams(raw transport.RawConnection, logger logging.Logger) *ResponseStreams {
	return &ResponseStreams{
		raw:     raw,
		streams: make(map[int]*ResponseStream),
		logger:  logger.New("response_streams"),
	}
}

func (r *ResponseStreams) Open(ctx context.Context, requestNumber int) (*ResponseStream, error) {
	r.streamsLock.Lock()
	defer r.streamsLock.Unlock()

	if r.closed {
		return nil, errors.New("response streams were closed")
	}

	if _, ok := r.streams[requestNumber]; ok {
		return nil, errors.New("response stream with this number already exists")
	}

	rs := NewResponseStream(ctx, requestNumber)
	r.streams[requestNumber] = rs

	go r.waitAndCloseResponseStream(rs)

	return rs, nil
}

func (r *ResponseStreams) HandleIncomingResponse(msg *transport.Message) error {
	if msg.Header.IsRequest() {
		return errors.New("passed a request")
	}

	r.streamsLock.Lock()
	defer r.streamsLock.Unlock()

	rs, ok := r.streams[-msg.Header.RequestNumber()]
	if !ok {
		return nil
	}

	var err error
	if msg.Header.Flags().EndOrError {
		err = ErrEndOrErr
		defer rs.cancel()
	}

	select {
	case rs.ch <- ResponseWithError{Value: NewResponse(msg.Body), Err: err}:
	case <-rs.ctx.Done():
	}

	return nil
}

func (s *ResponseStreams) Close() {
	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	s.closed = true

	for _, rs := range s.streams {
		rs.cancel()
	}
}

func (s *ResponseStreams) waitAndCloseResponseStream(rs *ResponseStream) {
	<-rs.ctx.Done()

	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	go func() {
		if err := s.sendCloseStream(rs.number); err != nil {
			s.logger.WithError(err).Debug("failed to close the stream")
		}
	}()

	delete(s.streams, rs.number)
	close(rs.ch)
}

func (s *ResponseStreams) sendCloseStream(number int) error {
	// todo do this correctly? are the flags correct?
	flags := transport.MessageHeaderFlags{
		Stream:     true,
		EndOrError: true,
		BodyType:   transport.MessageBodyTypeJSON,
	}

	j := []byte("true")

	header, err := transport.NewMessageHeader(
		flags,
		uint32(len(j)),
		int32(number),
	)
	if err != nil {
		return errors.Wrap(err, "could not create a message header")
	}

	msg, err := transport.NewMessage(header, j)
	if err != nil {
		return errors.Wrap(err, "could not create a message")
	}

	if err := s.raw.Send(&msg); err != nil {
		return errors.Wrap(err, "could not send a message")
	}

	return nil
}

type ResponseStream struct {
	number int
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan ResponseWithError
}

func NewResponseStream(ctx context.Context, number int) *ResponseStream {
	ctx, cancel := context.WithCancel(ctx)

	rs := &ResponseStream{
		number: number,
		ctx:    ctx,
		cancel: cancel,
		ch:     make(chan ResponseWithError),
	}

	return rs
}

func (rs ResponseStream) Channel() <-chan ResponseWithError {
	return rs.ch
}

type ResponseWithError struct {
	Value *Response
	Err   error
}
