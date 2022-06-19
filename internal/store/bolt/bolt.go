package bolt

import (
	"github.com/google/uuid"
	"go.etcd.io/bbolt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/store"
)

type BaseStore[T Entity] struct {
	Store      *bbolt.DB
	BucketName []byte

	marshalFn   func(T) []byte
	unmarshalFn func([]byte) (T, error)
}

type Entity interface {
	GetID() uuid.UUID
	GetEntityName() string
}

func NewBaseStore[T Entity](
	store *bbolt.DB,
	bucketName string,
	marshalFn func(T) []byte,
	unmarshalFn func([]byte) (T, error),
) *BaseStore[T] {

	return &BaseStore[T]{
		Store:       store,
		BucketName:  []byte(bucketName),
		marshalFn:   marshalFn,
		unmarshalFn: unmarshalFn,
	}
}

// Init creates a bucket if needed
func (base *BaseStore[T]) Init() (err error) {
	return base.Store.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(base.BucketName)
		if err != nil {
			return errors.Wrapf(err, "couldn't create bucket %q", base.BucketName)
		}
		return nil
	})
}

// GetByID returns an entity with passed id. If there's no such entity, the function returns NotFoundError
func (base *BaseStore[T]) GetByID(id uuid.UUID) (res T, err error) {
	err = base.Store.View(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket(base.BucketName)

		value := b.Get(id[:])
		if value == nil {
			return &store.NotFoundError{
				EntityName: base.getEntityName(),
				ID:         id,
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

// GetAll returns all entities in a bucket filtered with 'shouldSkipFn' and sorted with 'sortFn'.
func (base *BaseStore[T]) GetAll(shouldSkipFn func(T) bool, sortFn func([]T)) ([]T, error) {
	var res []T
	err := base.Store.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(base.BucketName)

		return b.ForEach(func(k, v []byte) error {
			entity, err := base.unmarshalFn(v)
			if err != nil {
				return err
			}
			if !shouldSkipFn(entity) {
				res = append(res, entity)
			}
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
func (base *BaseStore[T]) Create(entities ...T) error {
	return base.Store.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(base.BucketName)

		for _, entity := range entities {
			id := entity.GetID()
			data := base.marshalFn(entity)

			if b.Get(id[:]) != nil {
				return &store.AlreadyExistError{
					EntityName: base.getEntityName(),
					ID:         id,
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
func (base *BaseStore[T]) Update(entity T) error {
	id := entity.GetID()
	data := base.marshalFn(entity)

	return base.Store.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(base.BucketName)

		if b.Get(id[:]) == nil {
			return &store.NotFoundError{
				EntityName: base.getEntityName(),
				ID:         id,
			}
		}

		if err := b.Put(id[:], data); err != nil {
			return errors.Wrap(err, "put error")
		}
		return nil
	})
}

func (*BaseStore[T]) getEntityName() string {
	var zero T
	return zero.GetEntityName()
}
