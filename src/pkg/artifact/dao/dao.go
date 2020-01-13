// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"context"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/q"
)

// DAO is the data access object interface for artifact
type DAO interface {
	// Count returns the total count of artifacts according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List artifacts according to the query
	List(ctx context.Context, query *q.Query) (artifacts []*Artifact, err error)
	// Get the artifact specified by ID
	Get(ctx context.Context, id int64) (*Artifact, error)
	// Create the artifact
	Create(ctx context.Context, artifact *Artifact) (id int64, err error)
	// Delete the artifact specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Update updates the artifact. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, artifact *Artifact, props ...string) (err error)
	// CreateReference creates the artifact reference
	CreateReference(ctx context.Context, reference *ArtifactReference) (id int64, err error)
	// ListReferences lists the artifact references according to the query
	ListReferences(ctx context.Context, query *q.Query) (references []*ArtifactReference, err error)
	// DeleteReferences deletes the references referenced by the artifact specified by parent ID
	DeleteReferences(ctx context.Context, parentID int64) (err error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	if query != nil {
		// ignore the page number and size
		query = &q.Query{
			Keywords: query.Keywords,
		}
	}
	return orm.QuerySetter(ctx, &Artifact{}, query).Count()
}
func (d *dao) List(ctx context.Context, query *q.Query) ([]*Artifact, error) {
	artifacts := []*Artifact{}
	if _, err := orm.QuerySetter(ctx, &Artifact{}, query).All(&artifacts); err != nil {
		return nil, err
	}
	return artifacts, nil
}
func (d *dao) Get(ctx context.Context, id int64) (*Artifact, error) {
	artifact := &Artifact{
		ID: id,
	}
	if err := orm.GetOrmer(ctx).Read(artifact); err != nil {
		if e, ok := orm.IsNotFoundError(err, "artifact %d not found", id); ok {
			err = e
		}
		return nil, err
	}
	return artifact, nil
}
func (d *dao) Create(ctx context.Context, artifact *Artifact) (int64, error) {
	id, err := orm.GetOrmer(ctx).Insert(artifact)
	if e, ok := orm.IsConflictError(err, "artifact %s already exists under the repository %d",
		artifact.Digest, artifact.RepositoryID); ok {
		err = e
	}
	return id, err
}
func (d *dao) Delete(ctx context.Context, id int64) error {
	n, err := orm.GetOrmer(ctx).Delete(&Artifact{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return ierror.NotFoundError(nil).WithMessage("artifact %d not found", id)
	}
	return nil
}
func (d *dao) Update(ctx context.Context, artifact *Artifact, props ...string) error {
	n, err := orm.GetOrmer(ctx).Update(artifact, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return ierror.NotFoundError(nil).WithMessage("artifact %d not found", artifact.ID)
	}
	return nil
}
func (d *dao) CreateReference(ctx context.Context, reference *ArtifactReference) (int64, error) {
	id, err := orm.GetOrmer(ctx).Insert(reference)
	if e, ok := orm.IsConflictError(err, "reference already exists, parent artifact ID: %d, child artifact ID: %d",
		reference.ParentID, reference.ChildID); ok {
		err = e
	}
	return id, err
}
func (d *dao) ListReferences(ctx context.Context, query *q.Query) ([]*ArtifactReference, error) {
	references := []*ArtifactReference{}
	if _, err := orm.QuerySetter(ctx, &ArtifactReference{}, query).All(&references); err != nil {
		return nil, err
	}
	return references, nil
}
func (d *dao) DeleteReferences(ctx context.Context, parentID int64) error {
	// make sure the parent artifact exist
	_, err := d.Get(ctx, parentID)
	if err != nil {
		return err
	}
	_, err = orm.QuerySetter(ctx, &ArtifactReference{}, &q.Query{
		Keywords: map[string]interface{}{
			"parent_id": parentID,
		},
	}).Delete()
	return err
}