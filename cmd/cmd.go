// Package cmd contains application commands
package cmd

type Command interface {
	// Run is a blocking operation and should be called in a goroutine
	Run() error
	Shutdown()
}

type DefaultConfig struct {
	Version string
	GitHash string
}
