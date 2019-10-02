/*
Copyright (c) 2018 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package finder

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/vmware/govmomi/vapi/library"
)

// Finder is a helper object for finding content library objects by their
// inventory paths: /LIBRARY/ITEM/FILE.
//
// Wildcard characters `*` and `?` are both supported. However, the use
// of a wildcard character in the search string results in a full listing of
// that part of the path's server-side items.
//
// Path parts that do not use wildcard characters rely on server-side Find
// functions to find the path token by its name. Ironically finding one
// item with a direct path takes longer than if a wildcard is used because
// of the multiple round-trips. Direct paths will be more performant on
// systems that have numerous items.
type Finder struct {
	M *library.Manager
}

// NewFinder returns a new Finder.
func NewFinder(m *library.Manager) *Finder {
	return &Finder{m}
}

// Find finds one or more items that match the provided inventory path(s).
func (f *Finder) Find(
	ctx context.Context, ipath ...string) ([]FindResult, error) {

	if len(ipath) == 0 {
		ipath = []string{""}
	}
	var result []FindResult
	for _, p := range ipath {
		results, err := f.find(ctx, p)
		if err != nil {
			return nil, err
		}
		result = append(result, results...)
	}
	return result, nil
}

func (f *Finder) find(ctx context.Context, ipath string) ([]FindResult, error) {

	if ipath == "" {
		ipath = "*"
	}

	// Get the argument and remove any leading separator characters.
	ipath = strings.TrimPrefix(ipath, "/")

	// Tokenize the path into its distinct parts.
	parts := strings.Split(ipath, "/")

	// If there are more than three parts then the file name contains
	// the "/" character. In that case collapse any additional parts
	// back into the filename.
	if len(parts) > 3 {
		parts = []string{
			parts[0],
			parts[1],
			strings.Join(parts[2:], "/"),
		}
	}

	libs, err := f.findLibraries(ctx, parts[0])
	if err != nil {
		return nil, err
	}

	// If the path is a single token then the libraries are requested.
	if len(parts) < 2 {
		return libs, nil
	}

	items, err := f.findLibraryItems(ctx, libs, parts[1])
	if err != nil {
		return nil, err
	}

	// If the path is two tokens then the library items are requested.
	if len(parts) < 3 {
		return items, nil
	}

	// Get the library item files.
	return f.findLibraryItemFiles(ctx, items, parts[2])
}

// FindResult is the type of object returned from a Find operation.
type FindResult interface {

	// GetParent returns the parent of the find result. If the find result
	// is a Library then this function will return nil.
	GetParent() FindResult

	// GetPath returns the inventory path of the find result.
	GetPath() string

	// GetID returns the ID of the find result.
	GetID() string

	// GetName returns the name of the find result.
	GetName() string

	// GetResult gets the underlying library object.
	GetResult() interface{}
}

type findResult struct {
	result interface{}
	parent FindResult
}

func (f findResult) GetResult() interface{} {
	return f.result
}
func (f findResult) GetParent() FindResult {
	return f.parent
}
func (f findResult) GetPath() string {
	switch f.result.(type) {
	case library.Library:
		return fmt.Sprintf("/%s", f.GetName())
	case library.Item, library.File:
		return fmt.Sprintf("%s/%s", f.parent.GetPath(), f.GetName())
	default:
		return ""
	}
}

func (f findResult) GetID() string {
	switch t := f.result.(type) {
	case library.Library:
		return t.ID
	case library.Item:
		return t.ID
	default:
		return ""
	}
}

func (f findResult) GetName() string {
	switch t := f.result.(type) {
	case library.Library:
		return t.Name
	case library.Item:
		return t.Name
	case library.File:
		return t.Name
	default:
		return ""
	}
}

func (f findResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.GetResult())
}

func (f *Finder) findLibraries(
	ctx context.Context,
	token string) ([]FindResult, error) {

	if token == "" {
		token = "*"
	}

	var result []FindResult

	// If the token does not contain any wildcard characters then perform
	// a lookup by name using a server side call.
	if !strings.ContainsAny(token, "*?") {
		libIDs, err := f.M.FindLibrary(ctx, library.Find{Name: token})
		if err != nil {
			return nil, err
		}
		for _, id := range libIDs {
			lib, err := f.M.GetLibraryByID(ctx, id)
			if err != nil {
				return nil, err
			}
			result = append(result, findResult{result: *lib})
		}
		if len(result) == 0 {
			lib, err := f.M.GetLibraryByID(ctx, token)
			if err == nil {
				result = append(result, findResult{result: *lib})
			}
		}
		return result, nil
	}

	libs, err := f.M.GetLibraries(ctx)
	if err != nil {
		return nil, err
	}
	for _, lib := range libs {
		match, err := path.Match(token, lib.Name)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, findResult{result: lib})
		}
	}
	return result, nil
}

func (f *Finder) findLibraryItems(
	ctx context.Context,
	parents []FindResult, token string) ([]FindResult, error) {

	if token == "" {
		token = "*"
	}

	var result []FindResult

	for _, parent := range parents {
		// If the token does not contain any wildcard characters then perform
		// a lookup by name using a server side call.
		if !strings.ContainsAny(token, "*?") {
			childIDs, err := f.M.FindLibraryItems(
				ctx, library.FindItem{
					Name:      token,
					LibraryID: parent.GetID(),
				})
			if err != nil {
				return nil, err
			}
			for _, id := range childIDs {
				child, err := f.M.GetLibraryItem(ctx, id)
				if err != nil {
					return nil, err
				}
				result = append(result, findResult{
					result: *child,
					parent: parent,
				})
			}
			continue
		}

		children, err := f.M.GetLibraryItems(ctx, parent.GetID())
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			match, err := path.Match(token, child.Name)
			if err != nil {
				return nil, err
			}
			if match {
				result = append(
					result, findResult{parent: parent, result: child})
			}
		}
	}
	return result, nil
}

func (f *Finder) findLibraryItemFiles(
	ctx context.Context,
	parents []FindResult, token string) ([]FindResult, error) {

	if token == "" {
		token = "*"
	}

	var result []FindResult

	for _, parent := range parents {
		// If the token does not contain any wildcard characters then perform
		// a lookup by name using a server side call.
		if !strings.ContainsAny(token, "*?") {
			child, err := f.M.GetLibraryItemFile(ctx, parent.GetID(), token)
			if err != nil {
				return nil, err
			}
			result = append(result, findResult{
				result: *child,
				parent: parent,
			})
			continue
		}

		children, err := f.M.ListLibraryItemFiles(ctx, parent.GetID())
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			match, err := path.Match(token, child.Name)
			if err != nil {
				return nil, err
			}
			if match {
				result = append(
					result, findResult{parent: parent, result: child})
			}
		}
	}
	return result, nil
}
