package backup

type NoopBackuper struct {
	shutdownSignal chan struct{}
}

func NewNoopBackuper() *NoopBackuper {
	return &NoopBackuper{
		shutdownSignal: make(chan struct{}),
	}
}

func (b *NoopBackuper) Start() error {
	<-b.shutdownSignal
	return nil
}

func (b *NoopBackuper) Shutdown() error {
	close(b.shutdownSignal)
	return nil
}
