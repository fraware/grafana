package social

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/grafana/grafana/pkg/infra/remotecache"
	"github.com/grafana/grafana/pkg/login/social/models"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/setting"
)

func trueBoolPtr() *bool {
	b := true
	return &b
}

func falseBoolPtr() *bool {
	b := false
	return &b
}

func TestSocialAzureAD_UserInfo(t *testing.T) {
	type fields struct {
		providerCfg map[string]any
		cfg         *setting.Cfg
		usGovURL    bool
	}
	type args struct {
		client *http.Client
	}

	tests := []struct {
		name                     string
		fields                   fields
		claims                   *azureClaims
		args                     args
		settingAutoAssignOrgRole string
		want                     *models.BasicUserInfo
		wantErr                  bool
	}{
		{
			name: "Email in email claim",
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{},
				Name:              "My Name",
				ID:                "1234",
			},
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Viewer",
				Groups: []string{},
			},
		},
		{
			name: "No email",
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
			},
			claims: &azureClaims{
				Email:             "",
				PreferredUsername: "",
				Roles:             []string{},
				Name:              "My Name",
				ID:                "1234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "No id token",
			claims: nil,
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "US Government domain",
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{},
				Name:              "My Name",
				ID:                "1234",
			},
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
				usGovURL: true,
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Viewer",
				Groups: []string{},
			},
		},
		{
			name: "Email in preferred_username claim",
			claims: &azureClaims{
				Email:             "",
				PreferredUsername: "me@example.com",
				Roles:             []string{},
				Name:              "My Name",
				ID:                "1234",
			},
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Viewer",
				Groups: []string{},
			},
		},
		{
			name: "Admin role",
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"Admin"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Admin",
				Groups: []string{},
			},
		},
		{
			name: "Lowercase Admin role",
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"admin"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Admin",
				Groups: []string{},
			},
		},
		{
			name: "Only other roles",
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"AppAdmin"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Viewer",
				Groups: []string{},
			},
		},
		// TODO: @mgyongyosi check this test
		{
			name: "role from env variable",
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{},
				Name:              "My Name",
				ID:                "1234",
			},
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Editor",
				},
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Editor",
				Groups: []string{},
			},
		},
		{
			name: "Editor role",
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"Editor"},
				Name:              "My Name",
				ID:                "1234",
			},
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Editor",
				},
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Editor",
				Groups: []string{},
			},
		},
		{
			name: "Admin and Editor roles in claim",
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Editor",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"Admin", "Editor"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Admin",
				Groups: []string{},
			},
		},
		{
			name: "Grafana Admin but setting is disabled",
			fields: fields{
				providerCfg: map[string]any{
					"name":                       "azuread",
					"client_id":                  "client-id-example",
					"allow_assign_grafana_admin": false,
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Editor",
				},
			},

			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"GrafanaAdmin"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:             "1234",
				Name:           "My Name",
				Email:          "me@example.com",
				Login:          "me@example.com",
				Role:           "Admin",
				Groups:         []string{},
				IsGrafanaAdmin: nil,
			},
		},
		{
			name: "Editor roles in claim and GrafanaAdminAssignment enabled",
			fields: fields{
				providerCfg: map[string]any{
					"name":                       "azuread",
					"client_id":                  "client-id-example",
					"allow_assign_grafana_admin": true,
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"Editor"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:             "1234",
				Name:           "My Name",
				Email:          "me@example.com",
				Login:          "me@example.com",
				Role:           "Editor",
				Groups:         []string{},
				IsGrafanaAdmin: falseBoolPtr(),
			},
		},
		{
			name: "Grafana Admin and Editor roles in claim",
			fields: fields{
				providerCfg: map[string]any{
					"name":                       "azuread",
					"client_id":                  "client-id-example",
					"allow_assign_grafana_admin": true,
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"GrafanaAdmin", "Editor"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:             "1234",
				Name:           "My Name",
				Email:          "me@example.com",
				Login:          "me@example.com",
				Role:           "Admin",
				Groups:         []string{},
				IsGrafanaAdmin: trueBoolPtr(),
			},
		},
		{
			name: "Error if user is not a member of allowed_groups",
			fields: fields{
				providerCfg: map[string]any{
					"name":                       "azuread",
					"client_id":                  "client-id-example",
					"allow_assign_grafana_admin": false,
					"allowed_groups":             "dead-beef",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Editor",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{},
				Groups:            []string{"foo", "bar"},
				Name:              "My Name",
				ID:                "1234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error if user is not a member of allowed_organizations",
			fields: fields{
				providerCfg: map[string]any{
					"name":                       "azuread",
					"client_id":                  "client-id-example",
					"allow_assign_grafana_admin": false,
					"allowed_organizations":      "uuid-1234",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Editor",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				TenantID:          "uuid-5678",
				PreferredUsername: "",
				Roles:             []string{},
				Groups:            []string{"foo", "bar"},
				Name:              "My Name",
				ID:                "1234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No error if user is a member of allowed_organizations",
			fields: fields{
				providerCfg: map[string]any{
					"name":                  "azuread",
					"client_id":             "client-id-example",
					"allowed_organizations": "uuid-1234,uuid-5678",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				TenantID:          "uuid-5678",
				PreferredUsername: "",
				Roles:             []string{},
				Groups:            []string{"foo", "bar"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Viewer",
				Groups: []string{"foo", "bar"},
			},
			wantErr: false,
		},
		{
			name: "No Error if user is a member of allowed_groups",
			fields: fields{
				providerCfg: map[string]any{
					"name":                       "azuread",
					"client_id":                  "client-id-example",
					"allow_assign_grafana_admin": "false",
					"allowed_groups":             "foo, bar",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "Viewer",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{},
				Groups:            []string{"foo"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:     "1234",
				Name:   "My Name",
				Email:  "me@example.com",
				Login:  "me@example.com",
				Role:   "Viewer",
				Groups: []string{"foo"},
			},
		},
		{
			name: "Fetch groups when ClaimsNames and ClaimsSources is set",
			fields: fields{
				providerCfg: map[string]any{
					"name":      "azuread",
					"client_id": "client-id-example",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "",
				},
			},
			claims: &azureClaims{
				ID:                "1",
				Name:              "test",
				PreferredUsername: "test",
				Email:             "test@test.com",
				Roles:             []string{"Viewer"},
				ClaimNames:        claimNames{Groups: "src1"},
				ClaimSources:      nil, // set by the test
			},
			settingAutoAssignOrgRole: "",
			want: &models.BasicUserInfo{
				Id:     "1",
				Name:   "test",
				Email:  "test@test.com",
				Login:  "test@test.com",
				Role:   "Viewer",
				Groups: []string{"from_server"},
			},
			wantErr: false,
		},
		{
			name: "Fetch groups when forceUseGraphAPI is set",
			fields: fields{
				providerCfg: map[string]any{
					"name":                "azuread",
					"client_id":           "client-id-example",
					"force_use_graph_api": "true",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "",
				},
			},
			claims: &azureClaims{
				ID:                "1",
				Name:              "test",
				PreferredUsername: "test",
				Email:             "test@test.com",
				Roles:             []string{"Viewer"},
				ClaimNames:        claimNames{Groups: "src1"},
				ClaimSources:      nil,                    // set by the test
				Groups:            []string{"foo", "bar"}, // must be ignored
			},
			settingAutoAssignOrgRole: "",
			want: &models.BasicUserInfo{
				Id:     "1",
				Name:   "test",
				Email:  "test@test.com",
				Login:  "test@test.com",
				Role:   "Viewer",
				Groups: []string{"from_server"},
			},
			wantErr: false,
		},
		{
			name: "Fetch empty role when strict attribute role is true and no match",
			fields: fields{
				providerCfg: map[string]any{
					"name":                  "azuread",
					"client_id":             "client-id-example",
					"role_attribute_strict": "true",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"foo"},
				Groups:            []string{},
				Name:              "My Name",
				ID:                "1234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Fetch empty role when strict attribute role is true and no role claims returned",
			fields: fields{
				providerCfg: map[string]any{
					"name":                  "azuread",
					"client_id":             "client-id-example",
					"role_attribute_strict": "true",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole: "",
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{},
				Groups:            []string{},
				Name:              "My Name",
				ID:                "1234",
			},
			want:    nil,
			wantErr: true,
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Instantiate a signer using RSASSA-PSS (SHA256) with the given private key.
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS256, Key: privateKey}, (&jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]any{"kid": "1"},
	}).WithType("JWT"))
	require.NoError(t, err)

	// generate JWKS
	jwks := &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key:       privateKey.Public(),
				KeyID:     "1",
				Algorithm: "PS256",
				Use:       "sig",
			},
		},
	}

	authURL := "https://login.microsoftonline.com/1234/oauth2/v2.0/authorize"
	usGovAuthURL := "https://login.microsoftonline.us/1234/oauth2/v2.0/authorize"

	cache := remotecache.NewFakeCacheStorage()
	// put JWKS in cache
	jwksDump, err := json.Marshal(jwks)
	require.NoError(t, err)

	err = cache.Set(context.Background(), azureCacheKeyPrefix+"client-id-example", jwksDump, 0)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewAzureADProvider(tt.fields.providerCfg, tt.fields.cfg, featuremgmt.WithFeatures(), cache)
			require.NoError(t, err)

			if tt.fields.usGovURL {
				s.SocialBase.Endpoint.AuthURL = usGovAuthURL
			} else {
				s.SocialBase.Endpoint.AuthURL = authURL
			}

			cl := jwt.Claims{
				Audience:  jwt.Audience{"client-id-example"},
				Subject:   "subject",
				Issuer:    "issuer",
				NotBefore: jwt.NewNumericDate(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)),
			}

			var raw string
			if tt.claims != nil {
				tt.claims.Audience = "client-id-example"
				if tt.claims.ClaimNames.Groups != "" {
					server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						tokenParts := strings.Split(request.Header.Get("Authorization"), " ")
						require.Len(t, tokenParts, 2)
						require.Equal(t, "fake_token", tokenParts[1])

						writer.WriteHeader(http.StatusOK)

						type response struct {
							Value []string
						}
						res := response{Value: []string{"from_server"}}
						require.NoError(t, json.NewEncoder(writer).Encode(&res))
					}))
					// need to set the fake servers url as endpoint to capture request
					tt.claims.ClaimSources = map[string]claimSource{
						tt.claims.ClaimNames.Groups: {Endpoint: server.URL},
					}
				}
				raw, err = jwt.Signed(sig).Claims(cl).Claims(tt.claims).CompactSerialize()
				require.NoError(t, err)
			} else {
				raw, err = jwt.Signed(sig).Claims(cl).CompactSerialize()
				require.NoError(t, err)
			}

			token := &oauth2.Token{
				AccessToken: "fake_token",
			}
			if tt.claims != nil {
				token = token.WithExtra(map[string]any{"id_token": raw})
			}

			tt.args.client = s.Client(context.Background(), token)

			got, err := s.UserInfo(context.Background(), tt.args.client, token)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestSocialAzureAD_SkipOrgRole(t *testing.T) {
	type fields struct {
		SocialBase  *SocialBase
		providerCfg map[string]any
		cfg         *setting.Cfg
	}

	tests := []struct {
		name                     string
		fields                   fields
		claims                   *azureClaims
		settingAutoAssignOrgRole string
		want                     *models.BasicUserInfo
		wantErr                  bool
	}{
		{
			name: "Grafana Admin and Editor roles in claim, skipOrgRoleSync disabled should get roles, skipOrgRoleSyncBase disabled",
			fields: fields{
				providerCfg: map[string]any{
					"name":                       "azuread",
					"client_id":                  "client-id-example",
					"allow_assign_grafana_admin": "true",
					"skip_org_role_sync":         "false",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole:          "",
					OAuthSkipOrgRoleUpdateSync: false,
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"GrafanaAdmin", "Editor"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:             "1234",
				Name:           "My Name",
				Email:          "me@example.com",
				Login:          "me@example.com",
				Role:           "Admin",
				IsGrafanaAdmin: trueBoolPtr(),
				Groups:         []string{},
			},
		},
		{
			name: "Grafana Admin and Editor roles in claim, skipOrgRoleSync disabled should not get roles",
			fields: fields{
				providerCfg: map[string]any{
					"name":                       "azuread",
					"client_id":                  "client-id-example",
					"allow_assign_grafana_admin": "true",
					"skip_org_role_sync":         "false",
				},
				cfg: &setting.Cfg{
					AutoAssignOrgRole:          "",
					OAuthSkipOrgRoleUpdateSync: false,
				},
			},
			claims: &azureClaims{
				Email:             "me@example.com",
				PreferredUsername: "",
				Roles:             []string{"GrafanaAdmin", "Editor"},
				Name:              "My Name",
				ID:                "1234",
			},
			want: &models.BasicUserInfo{
				Id:             "1234",
				Name:           "My Name",
				Email:          "me@example.com",
				Login:          "me@example.com",
				Role:           "Admin",
				IsGrafanaAdmin: trueBoolPtr(),
				Groups:         []string{},
			},
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Instantiate a signer using RSASSA-PSS (SHA256) with the given private key.
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS256, Key: privateKey}, (&jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]any{"kid": "1"},
	}).WithType("JWT"))
	require.NoError(t, err)

	// generate JWKS
	jwks := &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key:       privateKey.Public(),
				KeyID:     "1",
				Algorithm: string(jose.PS256),
				Use:       "sig",
			},
		},
	}

	authURL := "https://login.microsoftonline.com/1234/oauth2/v2.0/authorize"
	cache := remotecache.NewFakeCacheStorage()
	// put JWKS in cache
	jwksDump, err := json.Marshal(jwks)
	require.NoError(t, err)

	err = cache.Set(context.Background(), azureCacheKeyPrefix+"client-id-example", jwksDump, 0)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewAzureADProvider(tt.fields.providerCfg, tt.fields.cfg, featuremgmt.WithFeatures(), cache)
			require.NoError(t, err)

			s.SocialBase.Endpoint.AuthURL = authURL

			cl := jwt.Claims{
				Subject:   "subject",
				Issuer:    "issuer",
				NotBefore: jwt.NewNumericDate(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)),
				Audience:  jwt.Audience{"leela", "fry"},
			}

			var raw string
			if tt.claims != nil {
				tt.claims.Audience = "client-id-example"
				if tt.claims.ClaimNames.Groups != "" {
					server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						tokenParts := strings.Split(request.Header.Get("Authorization"), " ")
						require.Len(t, tokenParts, 2)
						require.Equal(t, "fake_token", tokenParts[1])

						writer.WriteHeader(http.StatusOK)

						type response struct {
							Value []string
						}
						res := response{Value: []string{"from_server"}}
						require.NoError(t, json.NewEncoder(writer).Encode(&res))
					}))
					// need to set the fake servers url as endpoint to capture request
					tt.claims.ClaimSources = map[string]claimSource{
						tt.claims.ClaimNames.Groups: {Endpoint: server.URL},
					}
				}
				raw, err = jwt.Signed(sig).Claims(cl).Claims(tt.claims).CompactSerialize()
				require.NoError(t, err)
			} else {
				raw, err = jwt.Signed(sig).Claims(cl).CompactSerialize()
				require.NoError(t, err)
			}

			token := &oauth2.Token{
				AccessToken: "fake_token",
			}
			if tt.claims != nil {
				token = token.WithExtra(map[string]any{"id_token": raw})
			}

			provClient := s.Client(context.Background(), token)

			got, err := s.UserInfo(context.Background(), provClient, token)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestSocialAzureAD_InitializeExtraFields(t *testing.T) {
	type settingFields struct {
		forceUseGraphAPI     bool
		allowedOrganizations []string
	}
	testCases := []struct {
		name     string
		settings map[string]any
		want     settingFields
	}{
		{
			name: "forceUseGraphAPI is set to true",
			settings: map[string]any{
				"force_use_graph_api": "true",
			},
			want: settingFields{
				forceUseGraphAPI:     true,
				allowedOrganizations: []string{},
			},
		},
		{
			name: "allowedOrganizations is set",
			settings: map[string]any{
				"allowed_organizations": "uuid-1234,uuid-5678",
			},
			want: settingFields{
				forceUseGraphAPI:     false,
				allowedOrganizations: []string{"uuid-1234", "uuid-5678"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewAzureADProvider(tc.settings, &setting.Cfg{}, featuremgmt.WithFeatures(), nil)
			require.NoError(t, err)

			require.Equal(t, tc.want.forceUseGraphAPI, s.forceUseGraphAPI)
			require.Equal(t, tc.want.allowedOrganizations, s.allowedOrganizations)
		})
	}
}
