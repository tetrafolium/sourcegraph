package rules

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (GraphQLResolver) RulesInRuleContainer(ctx context.Context, containerID graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.RuleConnection, error) {
	// Check existence. TODO!(sqs) check perms, security
	container, err := GraphQLResolver{}.RuleContainerByID(ctx, containerID)
	if err != nil {
		return nil, err
	}
	dbContainer, err := toDBRuleContainer(container)
	if err != nil {
		return nil, err
	}

	opt := ruleConnectionArgsToListOptions(arg)
	opt.Container = dbContainer
	return &ruleConnection{opt: opt}, nil
}

func ruleConnectionArgsToListOptions(arg *graphqlutil.ConnectionArgs) dbRulesListOptions {
	var opt dbRulesListOptions
	arg.Set(&opt.LimitOffset)
	return opt
}

type ruleConnection struct {
	opt dbRulesListOptions

	once  sync.Once
	rules []*dbRule
	err   error
}

func (r *ruleConnection) compute(ctx context.Context) ([]*dbRule, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.rules, r.err = dbRules{}.List(ctx, opt2)
	})
	return r.rules, r.err
}

func (r *ruleConnection) Nodes(ctx context.Context) ([]graphqlbackend.Rule, error) {
	dbRules, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.opt.LimitOffset != nil && len(dbRules) > r.opt.LimitOffset.Limit {
		dbRules = dbRules[:r.opt.LimitOffset.Limit]
	}

	rules := make([]graphqlbackend.Rule, len(dbRules))
	for i, dbRule := range dbRules {
		rules[i] = newGQLRule(dbRule)
	}
	return rules, nil
}

func (r *ruleConnection) TotalCount(ctx context.Context) (int32, error) {
	count, err := dbRules{}.Count(ctx, r.opt)
	return int32(count), err
}

func (r *ruleConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	rules, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(rules) > r.opt.Limit), nil
}
