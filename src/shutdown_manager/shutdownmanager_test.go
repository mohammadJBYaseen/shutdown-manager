package shutdownmanager

import (
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
		<-dbShutdown // Wait for shutdown signalç

		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(2 * time.Millisecond) // Simulate cleanup work

		close(dbDone) // Signal completion
	}()

	// Simulate worker service
	go func() {
		// Service running...
		<-workerShutdown // Wait for shutdown signal

		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(1 * time.Millisecond) // Simulate cleanup work

		close(workerDone) // Signal completion
	}()

	// Simulate cache service
	go func() {
		// Service running...
		<-cacheShutdown // Wait for shutdown signal

		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work

		close(cacheDone) // Signal completion
	}()

	log.Debug().Msgf("Application started. Press Ctrl+C to shutdown.")
	go func ()  {
		time.Sleep(time.Second*1)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()
	// Handle shutdown
	err:= shutdownmanager.StartListner()
		
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
		<-dbShutdown // Wait for shutdown signalç

		log.Debug().Msgf("Database: Closing connections...")
		time.Sleep(600 * time.Millisecond) // Simulate cleanup work

		close(dbDone) // Signal completion
	}()

	// Simulate worker service
	go func() {
		// Service running...
		<-workerShutdown // Wait for shutdown signal

		log.Debug().Msgf("Worker: Finishing pending tasks...")
		time.Sleep(200 * time.Millisecond) // Simulate cleanup work

		close(workerDone) // Signal completion
	}()

	// Simulate cache service
	go func() {
		// Service running...
		<-cacheShutdown // Wait for shutdown signal

		log.Debug().Msgf("Cache: Flushing to disk...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup work

		close(cacheDone) // Signal completion
	}()

	log.Debug().Msgf("Application started. Press Ctrl+C to shutdown.")
	go func ()  {
		time.Sleep(time.Second*1)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()
	// Handle shutdown
	err:= shutdownmanager.StartListner()
		
	if err == nil {
		panic("shutdown manger should be timeouted")
	}
}
