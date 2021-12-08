package backup

import (
	"context"
	"os"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

type Backuper struct {
	cfg           Config
	db            Database
	log           logger.Logger
	backupManager *backupManager

	shutdownSignal chan struct{}
}

type Config struct {
	Dir         string
	Interval    time.Duration
	ExitOnError bool
}

type Database interface {
	Backup(ctx context.Context) ([]byte, error)
	GetType() db.Type
}

func NewBackuper(cfg Config, db Database, log logger.Logger) (*Backuper, error) {
	info, err := os.Stat(cfg.Dir)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't check dir %q", cfg.Dir)
	}
	if !info.IsDir() {
		return nil, errors.Errorf("%q is not a directory", cfg.Dir)
	}

	const minInterval = time.Minute
	if cfg.Interval < minInterval {
		return nil, errors.Errorf("minimum interval is %s, passed %s", minInterval, cfg.Interval)
	}

	return &Backuper{
		cfg:           cfg,
		db:            db,
		log:           log,
		backupManager: newBackupManager(cfg.Dir, db.GetType()),
		//
		shutdownSignal: make(chan struct{}),
	}, nil
}

func (b *Backuper) Start() error {
	ticker := time.NewTicker(b.cfg.Interval)
	defer ticker.Stop()

	defer close(b.shutdownSignal)

	backup := func(backupType string) (resErr error) {
		defer func() {
			const errorMsg = "couldn't backup data"
			if resErr != nil {
				if !b.cfg.ExitOnError {
					b.log.WithError(resErr).Error(errorMsg)
					resErr = nil
				}
				resErr = errors.Wrap(resErr, errorMsg)
			}
		}()

		b.log.Debugf("start %s backup", backupType)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		data, err := b.db.Backup(ctx)
		if err != nil {
			return err
		}

		backupFile, err := b.backupManager.NewBackupFile()
		if err != nil {
			return errors.Wrap(err, "couldn't create backup file")
		}
		if _, err := backupFile.Write(data); err != nil {
			backupFile.Close()

			return errors.Wrap(err, "couldn't write backup file")
		}
		if err := backupFile.Close(); err != nil {
			return errors.Wrap(err, "couldn't close backup file")
		}
		return nil
	}

	if err := backup("startup"); err != nil {
		return err
	}
	for {
		select {
		case <-ticker.C:
			if err := backup("scheduled"); err != nil {
				return err
			}

		case <-b.shutdownSignal:
			if err := backup("shutdown"); err != nil {
				return err
			}
			return nil
		}
	}
}

func (b *Backuper) Shutdown() error {
	select {
	case <-b.shutdownSignal:
	default:
		// Send a signal and wait for the actual shutdown
		b.shutdownSignal <- struct{}{}
		<-b.shutdownSignal
	}
	return nil
}
