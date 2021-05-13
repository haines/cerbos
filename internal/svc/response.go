// Copyright 2021 Zenauth Ltd.

package svc

import (
	enginev1 "github.com/cerbos/cerbos/internal/genpb/engine/v1"
	requestv1 "github.com/cerbos/cerbos/internal/genpb/request/v1"
	responsev1 "github.com/cerbos/cerbos/internal/genpb/response/v1"
	sharedv1 "github.com/cerbos/cerbos/internal/genpb/shared/v1"
)

type checkResourceSetResponseBuilder struct {
	*responsev1.CheckResourceSetResponse
	includeMeta bool
}

func newCheckResourceSetResponseBuilder(req *requestv1.CheckResourceSetRequest) *checkResourceSetResponseBuilder {
	resp := &responsev1.CheckResourceSetResponse{
		RequestId:         req.RequestId,
		ResourceInstances: make(map[string]*responsev1.CheckResourceSetResponse_ActionEffectMap, len(req.Resource.Instances)),
	}

	if req.IncludeMeta {
		resp.Meta = &responsev1.CheckResourceSetResponse_Meta{
			ResourceInstances: make(map[string]*responsev1.CheckResourceSetResponse_Meta_ActionMeta, len(req.Resource.Instances)),
		}
	}

	return &checkResourceSetResponseBuilder{
		CheckResourceSetResponse: resp,
		includeMeta:              req.IncludeMeta,
	}
}

func (resp *checkResourceSetResponseBuilder) addResult(resourceKey string, result *enginev1.CheckOutput) {
	actionsMap := make(map[string]sharedv1.Effect, len(result.Actions))
	for action, actionEffect := range result.Actions {
		actionsMap[action] = actionEffect.Effect
	}

	resp.ResourceInstances[resourceKey] = &responsev1.CheckResourceSetResponse_ActionEffectMap{
		Actions: actionsMap,
	}

	if resp.includeMeta {
		resp.addResultMeta(resourceKey, result)
	}
}

func (resp *checkResourceSetResponseBuilder) addResultMeta(resourceKey string, result *enginev1.CheckOutput) {
	meta := &responsev1.CheckResourceSetResponse_Meta_ActionMeta{
		Actions: make(map[string]*responsev1.CheckResourceSetResponse_Meta_EffectMeta, len(result.Actions)),
	}

	for action, actionEffect := range result.Actions {
		meta.Actions[action] = &responsev1.CheckResourceSetResponse_Meta_EffectMeta{
			MatchedPolicy: actionEffect.Policy,
		}
	}

	meta.EffectiveDerivedRoles = result.EffectiveDerivedRoles

	resp.Meta.ResourceInstances[resourceKey] = meta
}

func (resp *checkResourceSetResponseBuilder) build() *responsev1.CheckResourceSetResponse {
	return resp.CheckResourceSetResponse
}