package bolt

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"go.etcd.io/bbolt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/categories"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/store"
	"github.com/google/uuid"
)

const bucketName = "categories"

type Bolt struct {
	base *store.BaseBolt[categories.Category]
}

var _ categories.Store = (*Bolt)(nil)

func NewBolt(boltStore *bbolt.DB) (*Bolt, error) {
	store := &Bolt{
		base: store.NewBaseBolt(
			boltStore, bucketName, marshalBoltCategory, unmarshalBoltCategory,
		),
	}

	if err := store.base.Init(); err != nil {
		return nil, errors.Wrap(err, "couldn't init store")
	}
	return store, nil
}

func (bolt Bolt) GetAll(ctx context.Context) ([]categories.Category, error) {
	return bolt.base.GetAll(
		nil,
		func(categories []categories.Category) {
			sort.Slice(categories, func(i, j int) bool {
				if categories[i].ParentID == categories[j].ParentID {
					return categories[i].Name < categories[j].Name
				}
				return categories[i].ID.String() < categories[j].ID.String()
			})
		},
	)
}

func (bolt Bolt) Create(ctx context.Context, category categories.Category) error {
	return bolt.base.Create(category)
}

func (bolt Bolt) Update(ctx context.Context, category categories.Category) error {
	return bolt.base.Update(category)
}

type boltCategory struct {
	ID       uuid.UUID `json:"id"`
	ParentID uuid.UUID `json:"parent_id"`
	Name     string    `json:"name"`
}

func marshalBoltCategory(category categories.Category) []byte {
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

func unmarshalBoltCategory(data []byte) (categories.Category, error) {
	var category boltCategory
	if err := json.Unmarshal(data, &category); err != nil {
		return categories.Category{}, errors.Wrap(err, "couldn't unmarshal category")
	}

	return categories.Category{
		ID:       category.ID,
		ParentID: category.ParentID,
		Name:     category.Name,
	}, nil
}
