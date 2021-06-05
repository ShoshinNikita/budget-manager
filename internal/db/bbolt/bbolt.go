package bbolt

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

const (
	monthsBuckets         = "months"
	daysBucket            = "days"
	incomesBucket         = "incomes"
	monthlyPaymentsBucket = "monthly_payments"
	spendsBucket          = "spends"
	spendTypesBucket      = "spend_types"
)

// TODO
// There are 6 top-level buckets:
//
// 	* months
// 	* days
// 	* incomes
// 	* monthly_payments
// 	* spends
// 	* spend_types
//
type DB struct {
	db  *bbolt.DB
	log logrus.FieldLogger
}

type Config struct {
	Path string `env:"DB_BOLT_PATH" envDefault:"budget-manager.db"`
}

// NewDB opens a bolt file
func NewDB(config Config, log logrus.FieldLogger) (db *DB, err error) {
	db = &DB{
		log: log.WithField("db_type", "pg"),
	}
	db.db, err = bbolt.Open(config.Path, 0666, &bbolt.Options{
		Timeout: time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't open bolt file")
	}

	return db, nil
}

// Prepare creates missing buckets
func (db *DB) Prepare() error {
	return db.db.Update(func(tx *bbolt.Tx) error {
		for _, bucketName := range []string{
			monthsBuckets,
			daysBucket,
			incomesBucket,
			monthlyPaymentsBucket,
			spendsBucket,
			spendTypesBucket,
		} {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
				return errors.Wrapf(err, "couldn't create bucket '%s'", bucketName)
			}
		}
		return nil
	})
}

// Shutdown closes the connection to the db
func (db *DB) Shutdown() error {
	return db.db.Close()
}
