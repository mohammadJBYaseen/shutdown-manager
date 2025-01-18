package shutdownmanager

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	ShutdownManagerProperties struct {
		WaitTimeout     time.Duration
		ShutdownSignals []os.Signal
	}

	ShutdownManager interface {
		StartListner() error
	}

	ShutdownManagerWithCallBack interface {
		StartListner() error
		RegisterHook(func(context.Context) error)
	}

	ShutdownManagerWithSignals interface {
		StartListner() error
		RegisterService(string) (chan struct{}, chan struct{})
	}

	shutdownManagerWithCallBackImpl struct {
		hooks           []func(context.Context) error
		timeout         time.Duration
		waitGroup       sync.WaitGroup
		shutdownSignals []os.Signal
		started         bool
	}

	Service struct {
		name     string
		shutdown chan struct{}
		done     chan struct{}
	}

	shutdownManagerWithSignalsImp struct {
		services        []Service
		timeout         time.Duration
		shutdownSignals []os.Signal
		started        bool
	}
)

func NewShutdownManagerWithCallBack(ShutdownManagerProperties ShutdownManagerProperties) ShutdownManagerWithCallBack {
	return &shutdownManagerWithCallBackImpl{
		timeout:         ShutdownManagerProperties.WaitTimeout,
		shutdownSignals: ShutdownManagerProperties.ShutdownSignals,
		started:        false,
	}
}

func NewShutdownManagerWithSignals(ShutdownManagerProperties ShutdownManagerProperties) ShutdownManagerWithSignals {
	return &shutdownManagerWithSignalsImp{
		timeout:         ShutdownManagerProperties.WaitTimeout,
		shutdownSignals: ShutdownManagerProperties.ShutdownSignals,
		started:        false,
	}
}

func (sm *shutdownManagerWithCallBackImpl) RegisterHook(hook func(context.Context) error) {
	sm.hooks = append(sm.hooks, hook)
}

func (s *shutdownManagerWithCallBackImpl) StartListner() error {
	if s.started {
		return fmt.Errorf("shutdown manager already started")
	}
	s.started = true
	if len(s.hooks) <= 0 {
		log.Debug().Msgf("No Shutdown hooks is registered.")
		return nil
	}
	log.Debug().Msgf("Start Shutdown hooks.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, s.shutdownSignals...)

	sig := <-sigChan
	log.Debug().Msgf("Received shutdown signal: %v\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	for i, hook := range s.hooks {
		s.waitGroup.Add(1)
		go func(hookIndex int, hookFn func(context.Context) error) {
			defer s.waitGroup.Done()

			if err := hookFn(ctx); err != nil {
				log.Error().Msgf("Error in shutdown hook %d: %v\n", hookIndex, err)
			}
		}(i, hook)
	}

	done := make(chan struct{})
	go func() {
		s.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Debug().Msgf("All shutdown hooks completed successfully")
	case <-ctx.Done():
		log.Debug().Msgf("Global shutdown timeout exceeded")
		return fmt.Errorf("global shutdown timeout exceeded")
	}
	return nil
}

func (s *shutdownManagerWithSignalsImp) RegisterService(name string) (shutdown, done chan struct{}) {
	shutdown = make(chan struct{})
	done = make(chan struct{})

	s.services = append(s.services, Service{
		name:     name,
		shutdown: shutdown,
		done:     done,
	})

	return shutdown, done
}

func (s *shutdownManagerWithSignalsImp) fanIn(done ...chan struct{}) chan string {
	multiplexed := make(chan string)

	for i, ch := range done {
		go func(serviceIndex int, serviceDone chan struct{}) {
			<-serviceDone
			multiplexed <- s.services[serviceIndex].name
		}(i, ch)
	}

	return multiplexed
}

func (s *shutdownManagerWithSignalsImp) StartListner() error {
	if s.started {
		return fmt.Errorf("shutdown manager already started")
	}
	s.started = true
	if len(s.services) <= 0 {
		log.Debug().Msgf("No Shutdown hooks is registered.")
		return nil
	}

	log.Debug().Msgf("Start Shutdown hooks.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, s.shutdownSignals...)

	sig := <-sigChan
	log.Debug().Msgf("Received shutdown signal: %v\n", sig)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	for _, service := range s.services {
		log.Debug().Msgf("Signaling shutdown for service: %s\n", service.name)
		go func() {
			close(service.shutdown)
		}()
	}

	doneChannels := make([]chan struct{}, len(s.services))
	for i, service := range s.services {
		doneChannels[i] = service.done
	}

	multiplexed := s.fanIn(doneChannels...)

	completed := 0
	totalServices := len(s.services)

	for completed < totalServices {
		select {
		case serviceName := <-multiplexed:
			completed++
			log.Debug().Msgf("Service %s shutdown successfully (%d/%d complete)\n",
				serviceName, completed, totalServices)
		case <-ctx.Done():
			log.Debug().Msgf("Global shutdown timeout exceeded. Only %d/%d services completed\n",
				completed, totalServices)
			return fmt.Errorf("global shutdown timeout exceeded. Only %d/%d services completed",
				completed, totalServices)
		}
	}
	log.Debug().Msgf("All services shutdown successfully")
	return nil
}
