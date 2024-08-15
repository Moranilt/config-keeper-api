package callback

import (
	"context"
	"encoding/json"

	"github.com/Moranilt/config-keeper/pkg/file_contents"
	"github.com/Moranilt/config-keeper/pkg/files"
	"github.com/Moranilt/config-keeper/pkg/listeners"
	"github.com/Moranilt/http-utils/logger"
	"golang.org/x/sync/errgroup"
)

type CallbackService interface {
	// Run is the main loop of the callbackService. It listens for callback requests on the sendCh channel
	// and dispatches them to all registered listeners. The loop will continue until the provided context is canceled.
	Run(ctx context.Context)

	// prepareListenersData processes the callback request and prepares listener data
	// Returns a slice of listeners, serialized data, and any error encountered
	prepareListenersData(ctx context.Context, req *CallbackRequest) ([]*listeners.Listener, []byte, error)
}

type callbackService struct {
	log       logger.Logger
	sendCh    CallbackChannel
	file      files.Client
	listeners listeners.Client
	content   file_contents.Client
	rc        RequestsController
}

func New(
	log logger.Logger,
	sendCh CallbackChannel,
	file files.Client,
	listeners listeners.Client,
	content file_contents.Client,
	rc RequestsController,
) CallbackService {
	return &callbackService{
		log:       log,
		sendCh:    sendCh,
		file:      file,
		listeners: listeners,
		content:   content,
		rc:        rc,
	}
}

func (s *callbackService) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Stopping callback service")
			return
		case req := <-s.sendCh.Get():
			go func(req *CallbackRequest) {
				s.log.Infof("Received callback request with req %#v", req)
				listeners, fileData, err := s.prepareListenersData(ctx, req)
				if err != nil {
					s.log.Error("Error while preparing data")
					return
				}
				err = s.sendToListeners(ctx, listeners, fileData)
				if err != nil {
					s.log.Error("Error while sending data")
					return
				}
			}(req)
		}
	}
}

func (s *callbackService) prepareListenersData(ctx context.Context, req *CallbackRequest) ([]*listeners.Listener, []byte, error) {
	var err error
	s.log.Debugf("Request data for preparing listeners: %#v", req)
	file, err := s.file.Get(ctx, &files.GetRequest{ID: req.FileID})
	if err != nil {
		s.log.Errorf("Error getting file: %s", err)
		return nil, nil, err
	}

	fileContents, err := s.content.GetMany(ctx, &file_contents.GetManyRequest{FileID: req.FileID})
	if err != nil {
		s.log.Error("Error getting file contents: %s", err)
		return nil, nil, err
	}

	listenersList, err := s.listeners.GetMany(ctx, &listeners.GetManyRequest{FileID: req.FileID})
	if err != nil {
		s.log.Errorf("Error getting listeners: %s", err)
		return nil, nil, err
	}

	requestData := &FileData{
		File:        *file,
		FileContent: fileContents,
	}

	fileData, err := json.Marshal(requestData)
	if err != nil {
		s.log.Error("Error marshalling file: %s", err)
		return nil, nil, err
	}

	return listenersList, fileData, nil
}

func (s *callbackService) sendToListeners(ctx context.Context, listeners []*listeners.Listener, fileData []byte) error {
	g, ctx := errgroup.WithContext(ctx)
	limiter := make(chan struct{}, 10) // Limit to 10 concurrent requests

	for _, listener := range listeners {
		listener := listener
		limiter <- struct{}{}
		g.Go(func() error {
			defer func() { <-limiter }()
			return s.rc.SendRequestWithRetry(ctx, listener.CallbackEndpoint, fileData)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
