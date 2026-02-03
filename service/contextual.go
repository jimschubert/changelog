// Copyright 2020-2026 Jim Schubert
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/go-github/v29/github"
)

type clientContext struct{}
type Contextual struct {
	client *github.Client
}

func newContextual(client *github.Client) *Contextual {
	return &Contextual{client: client}
}

// CreateContext creates a known context from a parent context (c)
func (ctx *Contextual) CreateContext(c *context.Context) (context.Context, context.CancelFunc) {
	var parentContext context.Context
	if c != nil {
		parentContext = *c
	} else {
		parentContext = context.Background()
	}

	client := ctx.GetClient()
	if client != nil {
		parentContext = context.WithValue(parentContext, clientContext{}, client)
	}

	timeoutCtx, cancel := context.WithTimeout(parentContext, 10*time.Second)
	return timeoutCtx, cancel
}

// GetClient returns the client, if it exists
func (ctx *Contextual) GetClient() *github.Client {
	return ctx.client
}

// ClientFromContext returns a client if one exists on the context
func (ctx *Contextual) ClientFromContext(c *context.Context) (*github.Client, error) {
	client, ok := (*c).Value(clientContext{}).(*github.Client)
	if !ok {
		return nil, errors.New("client not available in this context")
	}

	return client, nil
}
