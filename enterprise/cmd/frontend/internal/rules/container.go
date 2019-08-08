package rules

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type ruleContainer struct {
	Campaign int64
	Thread   int64
}

func (c ruleContainer) graphqlID() graphql.ID {
	switch {
	case c.Campaign != 0:
		return graphqlbackend.MarshalCampaignID(c.Campaign)
	case c.Thread != 0:
		return graphqlbackend.MarshalThreadID(c.Thread)
	default:
		panic("invalid ruleContainer")
	}
}

func toDBRuleContainer(v *graphqlbackend.ToRuleContainer) (c ruleContainer, err error) {
	switch {
	case v.Campaign != nil:
		c.Campaign, err = graphqlbackend.UnmarshalCampaignID(v.Campaign.ID())
	case v.Thread != nil:
		c.Thread, err = graphqlbackend.UnmarshalThreadID(v.Thread.ID())
	default:
		panic("invalid ToRuleContainer")
	}
	return
}

func (GraphQLResolver) RuleContainerByID(ctx context.Context, id graphql.ID) (*graphqlbackend.ToRuleContainer, error) {
	node, err := graphqlbackend.NodeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var c graphqlbackend.ToRuleContainer
	switch relay.UnmarshalKind(id) {
	case graphqlbackend.GQLTypeCampaign:
		c.Campaign = node.(graphqlbackend.Campaign)
	case graphqlbackend.GQLTypeThread:
		c.Thread = node.(graphqlbackend.Thread)
	default:
		return nil, fmt.Errorf("node %q is not a RuleContainer", id)
	}
	return &c, nil
}
