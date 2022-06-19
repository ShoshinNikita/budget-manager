package bolt

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
)

type CategoriesStore struct {
	base *BaseStore[app.Category]
}

func NewCategoriesStore(boltStore *bbolt.DB) (*CategoriesStore, error) {
	store := &CategoriesStore{
		base: NewBaseStore(
			boltStore, "categories", marshalBoltCategory, unmarshalBoltCategory,
		),
	}

	if err := store.base.Init(); err != nil {
		return nil, errors.Wrap(err, "couldn't init store")
	}
	return store, nil
}

func (bolt CategoriesStore) GetAll(ctx context.Context) ([]app.Category, error) {
	return bolt.base.GetAll(
		nil,
		func(categories []app.Category) {
			sort.Slice(categories, func(i, j int) bool {
				if categories[i].ParentID == categories[j].ParentID {
					return categories[i].Name < categories[j].Name
				}
				return categories[i].ID.String() < categories[j].ID.String()
			})
		},
	)
}

func (bolt CategoriesStore) Create(ctx context.Context, category app.Category) error {
	return bolt.base.Create(category)
}

func (bolt CategoriesStore) Update(ctx context.Context, category app.Category) error {
	return bolt.base.Update(category)
}

type boltCategory struct {
	ID       uuid.UUID `json:"id"`
	ParentID uuid.UUID `json:"parent_id"`
	Name     string    `json:"name"`
}

func marshalBoltCategory(category app.Category) []byte {
	data, err := json.Marshal(boltCategory{
		ID:       category.ID,
		ParentID: category.ParentID,
		Name:     category.Name,
	})
	if err != nil {
		panic(fmt.Sprintf("error during category marshal: %s", err))
	}
	return data
}

func unmarshalBoltCategory(data []byte) (app.Category, error) {
	var category boltCategory
	if err := json.Unmarshal(data, &category); err != nil {
		return app.Category{}, errors.Wrap(err, "couldn't unmarshal category")
	}

	return app.Category{
		ID:       category.ID,
		ParentID: category.ParentID,
		Name:     category.Name,
	}, nil
}
