package anonimpl

import (
	"context"
	"net/http"
	"strings"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/anonymous"
	"github.com/grafana/grafana/pkg/services/authn"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/setting"
)

var _ authn.ContextAwareClient = new(Anonymous)

type Anonymous struct {
	cfg               *setting.Cfg
	log               log.Logger
	orgService        org.Service
	anonDeviceService anonymous.Service
}

func (a *Anonymous) Name() string {
	return authn.ClientAnonymous
}

func (a *Anonymous) Authenticate(ctx context.Context, r *authn.Request) (*authn.Identity, error) {
	o, err := a.orgService.GetByName(ctx, &org.GetOrgByNameQuery{Name: a.cfg.AnonymousOrgName})
	if err != nil {
		a.log.FromContext(ctx).Error("Failed to find organization", "name", a.cfg.AnonymousOrgName, "error", err)
		return nil, err
	}

	httpReqCopy := &http.Request{}
	if r.HTTPRequest != nil && r.HTTPRequest.Header != nil {
		// avoid r.HTTPRequest.Clone(context.Background()) as we do not require a full clone
		httpReqCopy.Header = r.HTTPRequest.Header.Clone()
		httpReqCopy.RemoteAddr = r.HTTPRequest.RemoteAddr
	}

	if err := a.anonDeviceService.TagDevice(ctx, httpReqCopy, anonymous.AnonDeviceUI); err != nil {
		a.log.Warn("Failed to tag anonymous session", "error", err)
	}

	return &authn.Identity{
		ID:           authn.AnonymousNamespaceID,
		OrgID:        o.ID,
		OrgName:      o.Name,
		OrgRoles:     map[int64]org.RoleType{o.ID: org.RoleType(a.cfg.AnonymousOrgRole)},
		ClientParams: authn.ClientParams{SyncPermissions: true},
	}, nil
}

func (a *Anonymous) Test(ctx context.Context, r *authn.Request) bool {
	// If anonymous client is register it can always be used for authentication
	return true
}

func (a *Anonymous) Priority() uint {
	return 100
}

func (a *Anonymous) UsageStatFn(ctx context.Context) (map[string]any, error) {
	m := map[string]any{}

	// Add stats about anonymous auth
	m["stats.anonymous.customized_role.count"] = 0
	if !strings.EqualFold(a.cfg.AnonymousOrgRole, "Viewer") {
		m["stats.anonymous.customized_role.count"] = 1
	}

	return m, nil
}
