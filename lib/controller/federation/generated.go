// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

package federation

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"

	"git.curoverse.com/arvados.git/sdk/go/arvados"
)

//
// -- this file is auto-generated -- do not edit -- edit list.go and run "go generate" instead --
//

func (conn *Conn) generated_ContainerList(ctx context.Context, options arvados.ListOptions) (arvados.ContainerList, error) {
	var mtx sync.Mutex
	var merged arvados.ContainerList
	var needSort atomic.Value
	needSort.Store(false)
	err := conn.splitListRequest(ctx, options, func(ctx context.Context, _ string, backend arvados.API, options arvados.ListOptions) ([]string, error) {
		cl, err := backend.ContainerList(ctx, options)
		if err != nil {
			return nil, err
		}
		mtx.Lock()
		defer mtx.Unlock()
		if len(merged.Items) == 0 {
			merged = cl
		} else if len(cl.Items) > 0 {
			merged.Items = append(merged.Items, cl.Items...)
			needSort.Store(true)
		}
		uuids := make([]string, 0, len(cl.Items))
		for _, item := range cl.Items {
			uuids = append(uuids, item.UUID)
		}
		return uuids, nil
	})
	if needSort.Load().(bool) {
		// Apply the default/implied order, "modified_at desc"
		sort.Slice(merged.Items, func(i, j int) bool {
			mi, mj := merged.Items[i].ModifiedAt, merged.Items[j].ModifiedAt
			return mj.Before(mi)
		})
	}
	if merged.Items == nil {
		// Return empty results as [], not null
		// (https://github.com/golang/go/issues/27589 might be
		// a better solution in the future)
		merged.Items = []arvados.Container{}
	}
	return merged, err
}

func (conn *Conn) generated_SpecimenList(ctx context.Context, options arvados.ListOptions) (arvados.SpecimenList, error) {
	var mtx sync.Mutex
	var merged arvados.SpecimenList
	var needSort atomic.Value
	needSort.Store(false)
	err := conn.splitListRequest(ctx, options, func(ctx context.Context, _ string, backend arvados.API, options arvados.ListOptions) ([]string, error) {
		cl, err := backend.SpecimenList(ctx, options)
		if err != nil {
			return nil, err
		}
		mtx.Lock()
		defer mtx.Unlock()
		if len(merged.Items) == 0 {
			merged = cl
		} else if len(cl.Items) > 0 {
			merged.Items = append(merged.Items, cl.Items...)
			needSort.Store(true)
		}
		uuids := make([]string, 0, len(cl.Items))
		for _, item := range cl.Items {
			uuids = append(uuids, item.UUID)
		}
		return uuids, nil
	})
	if needSort.Load().(bool) {
		// Apply the default/implied order, "modified_at desc"
		sort.Slice(merged.Items, func(i, j int) bool {
			mi, mj := merged.Items[i].ModifiedAt, merged.Items[j].ModifiedAt
			return mj.Before(mi)
		})
	}
	if merged.Items == nil {
		// Return empty results as [], not null
		// (https://github.com/golang/go/issues/27589 might be
		// a better solution in the future)
		merged.Items = []arvados.Specimen{}
	}
	return merged, err
}

func (conn *Conn) generated_UserList(ctx context.Context, options arvados.ListOptions) (arvados.UserList, error) {
	var mtx sync.Mutex
	var merged arvados.UserList
	err := conn.splitListRequest(ctx, options, func(ctx context.Context, _ string, backend arvados.API, options arvados.ListOptions) ([]string, error) {
		cl, err := backend.UserList(ctx, options)
		if err != nil {
			return nil, err
		}
		mtx.Lock()
		defer mtx.Unlock()
		if len(merged.Items) == 0 {
			merged = cl
		} else {
			merged.Items = append(merged.Items, cl.Items...)
		}
		uuids := make([]string, 0, len(cl.Items))
		for _, item := range cl.Items {
			uuids = append(uuids, item.UUID)
		}
		return uuids, nil
	})
	sort.Slice(merged.Items, func(i, j int) bool { return merged.Items[i].UUID < merged.Items[j].UUID })
	return merged, err
}
