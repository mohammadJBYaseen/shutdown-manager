package shutdownmanager

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

func Test_shutdownManagerWithSignalsImp_Shutdown_Successfully(t *testing.T) {

	shutdownmanager := NewShutdownManagerWithSignals(ShutdownManagerProperties{
		WaitTimeout:     time.Duration(time.Millisecond * 500),
		ShutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	})

	dbShutdown, dbDone := shutdownmanager.RegisterService("database")
	workerShutdown, workerDone := shutdownmanager.RegisterService("worker")
	cacheShutdown, cacheDone := shutdownmanager.RegisterService("cache")

	// Simulate database service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in shutdown handler %v: %v", "Database", r)
				log.Error().Msg(err.Error())
			}
			defer close(dbDone) // Signal completion
		}()
		<-dbShutdown // Wait for shutdown signalç

		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(2 * time.Millisecond) // Simulate cleanup work
	}()

	// Simulate worker service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in shutdown handler %v: %v", "Worker", r)
				log.Error().Msg(err.Error())
			}
			defer close(workerDone) // Signal completion
		}()
		// Service running...
		<-workerShutdown // Wait for shutdown signal

		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(1 * time.Millisecond) // Simulate cleanup work
	}()

	// Simulate cache service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in shutdown handler %v: %v", "Cache", r)
				log.Error().Msg(err.Error())
			}
			defer close(cacheDone) // Signal completion
		}()
		// Service running...
		<-cacheShutdown // Wait for shutdown signal

		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work
	}()

	log.Debug().Msgf("Application started.")
	go func() {
		time.Sleep(time.Second * 1)
		log.Debug().Msgf("Simulating SIG INT signal.")
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			panic("Can't send cancel signal")
		}
	}()
	// Handle shutdown
	err := shutdownmanager.StartListner()

	if err != nil {
		panic("shutdown manger not executed successfully")
	}
}

func Test_shutdownManagerWithSignalsImp_Shutdown_Failed_Timeouted(t *testing.T) {

	shutdownmanager := NewShutdownManagerWithSignals(ShutdownManagerProperties{
		WaitTimeout:     time.Duration(time.Millisecond * 500),
		ShutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	})

	dbShutdown, dbDone := shutdownmanager.RegisterService("database")
	workerShutdown, workerDone := shutdownmanager.RegisterService("worker")
	cacheShutdown, cacheDone := shutdownmanager.RegisterService("cache")

	// Simulate database service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in shutdown handler %v: %v", "Database", r)
				log.Error().Msg(err.Error())
			}
			defer close(dbDone) // Signal completion
		}()
		<-dbShutdown // Wait for shutdown signalç

		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(600 * time.Millisecond) // Simulate cleanup work
	}()

	// Simulate worker service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in shutdown handler %v: %v", "Worker", r)
				log.Error().Msg(err.Error())
			}
			defer close(workerDone) // Signal completion
		}()
		// Service running...
		<-workerShutdown // Wait for shutdown signal

		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(200 * time.Millisecond) // Simulate cleanup work
	}()

	// Simulate cache service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in shutdown handler %v: %v", "Cache", r)
				log.Error().Msg(err.Error())
			}
			defer close(cacheDone) // Signal completion
		}()
		// Service running...
		<-cacheShutdown // Wait for shutdown signal

		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work
	}()

	log.Debug().Msgf("Application started.")
	go func() {
		time.Sleep(time.Second * 1)
		log.Debug().Msgf("Simulating SIG INT signal.")
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			panic("Can't send cancel signal")
		}
	}()
	// Handle shutdown
	err := shutdownmanager.StartListner()

	if err == nil {
		panic("shutdown manger should be timeouted")
	}
}


func Test_shutdownManagerWithSignalsImp_Shutdown_Successfully_Even_Hook_Failed_For_UnexpectedErrors(t *testing.T) {

	shutdownmanager := NewShutdownManagerWithSignals(ShutdownManagerProperties{
		WaitTimeout:     time.Duration(time.Millisecond * 500),
		ShutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	})


	dbShutdown, dbDone := shutdownmanager.RegisterService("database")
	workerShutdown, workerDone := shutdownmanager.RegisterService("worker")
	cacheShutdown, cacheDone := shutdownmanager.RegisterService("cache")

	// Simulate database service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in shutdown handler %v: %v", "Database", r)
				log.Error().Msg(err.Error())
			}
			defer close(dbDone) // Signal completion
		}()
		<-dbShutdown // Wait for shutdown signalç

		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(2 * time.Millisecond) // Simulate cleanup work
		panic("Database: Closing connections failed")
	}()

	// Simulate worker service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in shutdown handler %v: %v", "Worker", r)
				log.Error().Msg(err.Error())
			}
			defer close(workerDone) // Signal completion
		}()
		// Service running...
		<-workerShutdown // Wait for shutdown signal

		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(1 * time.Millisecond) // Simulate cleanup work
	}()

	// Simulate cache service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in shutdown handler %v: %v", "Cache", r)
				log.Error().Msg(err.Error())
			}
			defer close(cacheDone) // Signal completion
		}()
		// Service running...
		<-cacheShutdown // Wait for shutdown signal

		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work

	}()

	log.Debug().Msgf("Application started.")
	go func() {
		time.Sleep(time.Second * 1)
		log.Debug().Msgf("Simulating SIG INT signal.")
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			panic("Can't send cancel signal")
		}
	}()
	// Handle shutdown
	err := shutdownmanager.StartListner()

	if err != nil {
		panic("shutdown manger should be completed")
	}
}

func Test_shutdownManagerWithSignalsImp_Shutdown_Successfully_Even_Hook_Failed(t *testing.T) {

	shutdownmanager := NewShutdownManagerWithCallBack(ShutdownManagerProperties{
		WaitTimeout:     time.Duration(time.Millisecond * 500),
		ShutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	})

	// Simulate database service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(50 * time.Millisecond) // Simulate cleanup work
		return fmt.Errorf("Failed to close Database connections ")
	})

	// Simulate worker service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(200 * time.Millisecond) // Simulate cleanup work
		return fmt.Errorf("Failed to close worker")
	})

	// Simulate cache service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work
		return nil
	})

	log.Debug().Msgf("Application started.")
	go func() {
		time.Sleep(time.Second * 1)
		log.Debug().Msgf("Simulating SIG INT signal.")
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			panic("Can't send cancel signal")
		}
	}()
	// Handle shutdown
	err := shutdownmanager.StartListner()

	if err != nil {
		panic("shutdown manger should be completed")
	}
}

func Test_shutdownManagerWithCallBackImpl_Shutdown_Successfully(t *testing.T) {

	shutdownmanager := NewShutdownManagerWithCallBack(ShutdownManagerProperties{
		WaitTimeout:     time.Duration(time.Millisecond * 500),
		ShutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	})

	// Simulate database service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(2 * time.Millisecond) // Simulate cleanup work
		return nil
	})

	// Simulate worker service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(1 * time.Millisecond) // Simulate cleanup work
		return nil
	})

	// Simulate cache service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work
		return nil
	})

	log.Debug().Msgf("Application started.")
	go func() {
		time.Sleep(time.Second * 1)
		log.Debug().Msgf("Simulating SIG INT signal.")
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			panic("Can't send cancel signal")
		}
	}()
	// Handle shutdown
	err := shutdownmanager.StartListner()

	if err != nil {
		panic("shutdown manger not executed successfully")
	}
}

func Test_shutdownManagerWithCallBackImpl_Shutdown_Failed_Timeouted(t *testing.T) {

	shutdownmanager := NewShutdownManagerWithCallBack(ShutdownManagerProperties{
		WaitTimeout:     time.Duration(time.Millisecond * 500),
		ShutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	})

	// Simulate database service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(600 * time.Millisecond) // Simulate cleanup work
		return nil
	})

	// Simulate worker service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(200 * time.Millisecond) // Simulate cleanup work
		return nil
	})

	// Simulate cache service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work
		return nil
	})

	log.Debug().Msgf("Application started.")
	go func() {
		time.Sleep(time.Second * 1)
		log.Debug().Msgf("Simulating SIG INT signal.")
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			panic("Can't send cancel signal")
		}
	}()
	// Handle shutdown
	err := shutdownmanager.StartListner()

	if err == nil {
		panic("shutdown manger should be timeouted")
	}
}

func Test_shutdownManagerWithCallBackImpl_Shutdown_Successfully_Event_Hook_Failed_For_UnexpectedErrors(t *testing.T) {

	shutdownmanager := NewShutdownManagerWithCallBack(ShutdownManagerProperties{
		WaitTimeout:     time.Duration(time.Millisecond * 500),
		ShutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	})

	// Simulate database service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(50 * time.Millisecond) // Simulate cleanup work
		return fmt.Errorf("Failed to close Database connections ")
	})

	// Simulate worker service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(200 * time.Millisecond) // Simulate cleanup work
		panic("Failed to close worker")
	})

	// Simulate cache service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work
		return nil
	})

	log.Debug().Msgf("Application started.")
	go func() {
		time.Sleep(time.Second * 1)
		log.Debug().Msgf("Simulating SIG INT signal.")
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			panic("Can't send cancel signal")
		}
	}()
	// Handle shutdown
	err := shutdownmanager.StartListner()

	if err != nil {
		panic("shutdown manger should be completed")
	}
}

func Test_shutdownManagerWithCallBackImpl_Shutdown_Successfully_Event_Hook_Failed(t *testing.T) {

	shutdownmanager := NewShutdownManagerWithCallBack(ShutdownManagerProperties{
		WaitTimeout:     time.Duration(time.Millisecond * 500),
		ShutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	})

	// Simulate database service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(50 * time.Millisecond) // Simulate cleanup work
		return fmt.Errorf("Failed to close Database connections ")
	})

	// Simulate worker service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(200 * time.Millisecond) // Simulate cleanup work
		return fmt.Errorf("Failed to close worker")
	})

	// Simulate cache service
	shutdownmanager.RegisterHook(func(ctx context.Context) error {
		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work
		return nil
	})

	log.Debug().Msgf("Application started.")
	go func() {
		time.Sleep(time.Second * 1)
		log.Debug().Msgf("Simulating SIG INT signal.")
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			panic("Can't send cancel signal")
		}
	}()
	// Handle shutdown
	err := shutdownmanager.StartListner()

	if err != nil {
		panic("shutdown manger should be completed")
	}
}
