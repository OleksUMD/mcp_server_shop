package jsonrpc

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Handler func(params json.RawMessage) (any, error)

type Server struct {
	handlers map[string]Handler
	logger   *logrus.Logger
}

func NewServer(logger *logrus.Logger) *Server {
	return &Server{
		handlers: make(map[string]Handler),
		logger:   logger,
	}
}

func (s *Server) RegisterMethod(method string, handler Handler) {
	s.handlers[method] = handler
	s.logger.WithField("method", method).Debug("Registered handler")
}

func (s *Server) HandleRequest(req *Request) *Response {
	s.logger.WithFields(logrus.Fields{
		"method": req.Method,
		"id":     req.ID,
	}).Debug("Handling request")
	if err := req.Validate(); err != nil {
		var jsonErr *Error
		if errors.As(err, &jsonErr) {
			return NewErrorResponse(req.ID, jsonErr)
		}
		return NewErrorResponse(req.ID, NewInternalError(err))
	}

	handler, ok := s.handlers[req.Method]
	if !ok {
		return NewErrorResponse(req.ID, NewMethodNotFoundError(req.Method))
	}

	result, err := handler(req.Params)
	if err != nil {
		var jsonErr *Error
		if errors.As(err, &jsonErr) {
			return NewErrorResponse(req.ID, jsonErr)
		}
		return NewErrorResponse(req.ID, NewInternalError(err))
	}

	return NewSuccessResponse(req.ID, result)
}

func (s *Server) ServeStdio() error {
	s.logger.Info("Starting server on stdio")
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				s.logger.Info("EOF received, shutting down server")
				return nil
			}
			s.logger.WithError(err).Error("unable to read from stdin")
			return err
		}

		s.logger.WithField("request", string(line)).Debug("Received request")
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			res := NewErrorResponse(nil, NewParseError(err))
			s.writeResponse(writer, res)
			continue
		}
		res := s.HandleRequest(&req)
		if !req.IsNotification() {
			s.writeResponse(writer, res)
		}
	}
}

func (s *Server) writeResponse(writer *bufio.Writer, res *Response) {
	bs, err := json.Marshal(res)
	if err != nil {
		s.logger.WithError(err).Error("unable to marshal response")
		return
	}
	s.logger.WithField("response", string(bs)).Debug("Sending response")
	writer.Write(bs)
	writer.WriteByte('\n')
	writer.Flush()
}
