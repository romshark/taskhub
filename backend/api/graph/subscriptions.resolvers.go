package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.34

import (
	"context"
	"time"

	"github.com/romshark/taskhub/api/auth"
	"github.com/romshark/taskhub/api/graph/model"
	"github.com/romshark/taskhub/api/reqctx"
	"golang.org/x/exp/slog"
)

// TaskUpsert is the resolver for the taskUpsert field.
func (r *subscriptionResolver) TaskUpsert(ctx context.Context) (<-chan *model.Task, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	c := make(chan *model.Task, 1)
	r.broadcastTaskUpsert.Subscribe(ctx, c)
	go logSubscriptionTermination(ctx, "taskUpsert")
	return c, nil
}

// ProjectUpsert is the resolver for the projectUpsert field.
func (r *subscriptionResolver) ProjectUpsert(ctx context.Context) (<-chan *model.Project, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	c := make(chan *model.Project, 1)
	r.broadcastProjectUpsert.Subscribe(ctx, c)
	go logSubscriptionTermination(ctx, "projectUpsert")
	return c, nil
}

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type subscriptionResolver struct{ *Resolver }

func logSubscriptionTermination(ctx context.Context, name string) {
	<-ctx.Done()
	c := reqctx.GetRequestContext(ctx)
	c.Log.Info(
		"subscription terminated",
		slog.String("requestID", string(c.RequestID)),
		slog.String("name", name),
		slog.String("duration", time.Since(c.Start).String()),
	)
}
