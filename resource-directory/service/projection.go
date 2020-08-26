package service

import (
	"context"
	"fmt"
	"time"

	projectionRA "github.com/go-ocf/cloud/resource-aggregate/cqrs/projection"
	"github.com/go-ocf/cqrs/eventbus"
	"github.com/go-ocf/cqrs/eventstore"
	"github.com/go-ocf/kit/log"
	"github.com/go-ocf/kit/strings"
	cache "github.com/patrickmn/go-cache"
)

// hasMatchingType returns true for matching a resource type.
// An empty typeFilter matches all resource types.
func hasMatchingType(resourceTypes []string, typeFilter strings.Set) bool {
	if len(typeFilter) == 0 {
		return true
	}
	if len(resourceTypes) == 0 {
		return false
	}
	return typeFilter.HasOneOf(resourceTypes...)
}

type Projection struct {
	projection *projectionRA.Projection
	cache      *cache.Cache
}

func NewProjection(ctx context.Context, name string, store eventstore.EventStore, subscriber eventbus.Subscriber, expiration time.Duration) (*Projection, error) {
	projection, err := projectionRA.NewProjection(ctx, name, store, subscriber, NewResourceCtx())
	if err != nil {
		return nil, fmt.Errorf("cannot create server: %v", err)
	}
	cache := cache.New(expiration, expiration)
	cache.OnEvicted(func(deviceId string, _ interface{}) {
		log.Debugf("device %v was remove from the cache", deviceId)
		projection.Unregister(deviceId)
	})
	return &Projection{projection: projection, cache: cache}, nil
}

func (p *Projection) GetResourceCtxs(ctx context.Context, resourceIdsFilter, typeFilter, deviceIds strings.Set) (map[string]map[string]*resourceCtx, error) {
	models := make([]eventstore.Model, 0, 32)
	for deviceId := range deviceIds {
		created, err := p.projection.Register(ctx, deviceId)
		if err != nil {
			return nil, fmt.Errorf("cannot register to projection %v", err)
		}
		if !created {
			defer func(ID string) {
				p.projection.Unregister(ID)
			}(deviceId)
		}
		p.cache.Set(deviceId, created, cache.DefaultExpiration)
		if len(resourceIdsFilter) > 0 {
			for resourceId := range resourceIdsFilter {
				m := p.projection.Models(deviceId, resourceId)
				if len(m) == 0 {
					err = p.projection.ForceUpdate(ctx, deviceId, resourceId)
					if err == nil {
						m = p.projection.Models(deviceId, resourceId)
					}
				}
				if len(m) > 0 {
					models = append(models, m...)
				}
			}
		} else {
			m := p.projection.Models(deviceId, "")
			if len(m) == 0 {
				err = p.projection.ForceUpdate(ctx, deviceId, "")
				if err == nil {
					m = p.projection.Models(deviceId, "")
				}
			}
			if len(m) > 0 {
				models = append(models, m...)
			}
		}
	}

	clonedModels := make(map[string]map[string]*resourceCtx)
	for _, m := range models {
		model := m.(*resourceCtx).Clone()
		if !model.snapshot.IsPublished {
			continue
		}
		if !hasMatchingType(model.snapshot.Resource.ResourceTypes, typeFilter) {
			continue
		}
		resources, ok := clonedModels[model.snapshot.GroupId()]
		if !ok {
			resources = make(map[string]*resourceCtx)
			clonedModels[model.snapshot.GroupId()] = resources
		}
		resources[model.snapshot.AggregateId()] = model
	}

	return clonedModels, nil
}