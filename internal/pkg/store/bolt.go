package store

import (
	"github.com/google/uuid"
	"go.etcd.io/bbolt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
)

type BaseBolt[T Entity] struct {
	Store      *bbolt.DB
	BucketName []byte

	marshalFn   func(T) []byte
	unmarshalFn func([]byte) (T, error)
}

type Entity interface {
	GetID() uuid.UUID
}

func NewBaseBolt[T Entity](
	store *bbolt.DB,
	bucketName string,
	marshalFn func(T) []byte,
	unmarshalFn func([]byte) (T, error),
) *BaseBolt[T] {

	return &BaseBolt[T]{
		Store:       store,
		BucketName:  []byte(bucketName),
		marshalFn:   marshalFn,
		unmarshalFn: unmarshalFn,
	}
}

// Init creates a bucket if needed
func (base *BaseBolt[T]) Init() (err error) {
	return base.Store.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(base.BucketName)
		return errors.Wrapf(err, "couldn't create bucket %q", base.BucketName)
	})
}

// GetByID returns an entity with passed id. If there's to such entity, it returns NotFoundError
func (base *BaseBolt[T]) GetByID(id uuid.UUID) (res T, err error) {
	err = base.Store.View(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket(base.BucketName)

		value := b.Get(id[:])
		if value == nil {
			return NotFoundError{
				Source: string(base.BucketName),
				ID:     id,
			}
		}
		res, err = base.unmarshalFn(value)
		if err != nil {
			return err
		}
		return nil
	})
	return res, err
}

// GetAll returns all entities in a bucket filtered with 'filterFn' and sorted with 'sortFn'.
func (base *BaseBolt[T]) GetAll(filterFn func(T) bool, sortFn func([]T)) ([]T, error) {
	var res []T
	err := base.Store.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(base.BucketName)

		return b.ForEach(func(k, v []byte) error {
			entity, err := base.unmarshalFn(v)
			if err != nil {
				return err
			}
			res = append(res, entity)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	if sortFn != nil {
		sortFn(res)
	}

	return res, nil
}

// Create saves the marshalled entities using their ids as the keys.
// It returns AlreadyExistError if a caller tries to overwrite an existing value.
func (base *BaseBolt[T]) Create(entities ...T) error {
	return base.Store.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(base.BucketName)

		for _, entity := range entities {
			id := entity.GetID()
			data := base.marshalFn(entity)

			if b.Get(id[:]) != nil {
				return AlreadyExistError{
					Source: string(base.BucketName),
					ID:     id,
				}
			}

			if err := b.Put(id[:], data); err != nil {
				return errors.Wrapf(err, "put error for entity %#v", entity)
			}
		}
		return nil
	})
}

// Update overwrites the value for entity id. It returns NotFoundError if there's no previous value.
func (base *BaseBolt[T]) Update(entity T) error {
	id := entity.GetID()
	data := base.marshalFn(entity)

	return base.Store.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(base.BucketName)

		if b.Get(id[:]) == nil {
			return NotFoundError{
				Source: string(base.BucketName),
				ID:     id,
			}
		}

		if err := b.Put(id[:], data); err != nil {
			return errors.Wrap(err, "put error")
		}
		return nil
	})
}
