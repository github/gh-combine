package github

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

type Client struct {
	REST    *api.RESTClient
	GraphQL *api.GraphQLClient
}

type PR struct {
	Number int
	Title  string
	Head   struct {
		Ref string
	}
	Base struct {
		Ref string
		SHA string
	}
	Labels []struct {
		Name string
	}
}

func New() (Client, error) {
	var err error

	c := Client{}

	if c.REST, err = api.DefaultRESTClient(); err != nil {
		return Client{}, fmt.Errorf("failed to get the default REST client: %w", err)
	}

	if c.GraphQL, err = api.DefaultGraphQLClient(); err != nil {
		return Client{}, fmt.Errorf("failed to create the default GraphQL Client : %w", err)
	}

	return c, nil
}

func (c Client) GetOpenPRs(ctx context.Context, repo Repo) ([]PR, error) {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue processing
	}

	var pulls []PR

	endpoint := fmt.Sprintf("repos/%s/%s/pulls?state=open", owner, repoName)
	if err := c.REST.DoWithContext(ctx, "GET", endpoint, &pulls); err != nil {
		return err
	}
}
