package portal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/workos-inc/workos-go/pkg/common"
)

func TestListOrganizations(t *testing.T) {
	tests := []struct {
		scenario string
		client   *Client
		options  ListOrganizationsOpts
		expected ListOrganizationsResponse
		err      bool
	}{
		{
			scenario: "Request without API Key returns an error",
			client:   &Client{},
			err:      true,
		},
		{
			scenario: "Request returns Organizations",
			client: &Client{
				APIKey: "test",
			},
			options: ListOrganizationsOpts{
				Domains: []string{"foo-corp.com"},
			},

			expected: ListOrganizationsResponse{
				Data: []Organization{
					Organization{
						ID:   "organization_id",
						Name: "Foo Corp",
						Domains: []OrganizationDomain{
							OrganizationDomain{
								ID:     "organization_domain_id",
								Domain: "foo-corp.com",
							},
						},
					},
				},
				ListMetadata: common.ListMetadata{
					Before: "",
					After:  "",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(listOrganizationsTestHandler))
			defer server.Close()

			client := test.client
			client.Endpoint = server.URL
			client.HTTPClient = server.Client()

			organizations, err := client.ListOrganizations(context.Background(), test.options)
			if test.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, organizations)
		})
	}
}

func listOrganizationsTestHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "Bearer test" {
		http.Error(w, "bad auth", http.StatusUnauthorized)
		return
	}

	if userAgent := r.Header.Get("User-Agent"); !strings.Contains(userAgent, "workos-go/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(struct {
		ListOrganizationsResponse
	}{
		ListOrganizationsResponse: ListOrganizationsResponse{
			Data: []Organization{
				Organization{
					ID:   "organization_id",
					Name: "Foo Corp",
					Domains: []OrganizationDomain{
						OrganizationDomain{
							ID:     "organization_domain_id",
							Domain: "foo-corp.com",
						},
					},
				},
			},
			ListMetadata: common.ListMetadata{
				Before: "",
				After:  "",
			},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func TestCreateOrganization(t *testing.T) {
	tests := []struct {
		scenario string
		client   *Client
		options  CreateOrganizationOpts
		expected Organization
		err      bool
	}{
		{
			scenario: "Request without API Key returns an error",
			client:   &Client{},
			err:      true,
		},
		{
			scenario: "Request returns Organization",
			client: &Client{
				APIKey: "test",
			},
			options: CreateOrganizationOpts{
				Name:    "Foo Corp",
				Domains: []string{"foo-corp.com"},
			},
			expected: Organization{
				ID:   "organization_id",
				Name: "Foo Corp",
				Domains: []OrganizationDomain{
					OrganizationDomain{
						ID:     "organization_domain_id",
						Domain: "foo-corp.com",
					},
				},
			},
		},
		{
			scenario: "Request with duplicate Organization Domain returns error",
			client: &Client{
				APIKey: "test",
			},
			err: true,
			options: CreateOrganizationOpts{
				Name:    "Foo Corp",
				Domains: []string{"duplicate.com"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(createOrganizationTestHandler))
			defer server.Close()

			client := test.client
			client.Endpoint = server.URL
			client.HTTPClient = server.Client()

			organization, err := client.CreateOrganization(context.Background(), test.options)
			if test.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, organization)
		})
	}
}

func createOrganizationTestHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "Bearer test" {
		http.Error(w, "bad auth", http.StatusUnauthorized)
		return
	}

	var opts CreateOrganizationOpts
	json.NewDecoder(r.Body).Decode(&opts)
	for _, domain := range opts.Domains {
		if domain == "duplicate.com" {
			http.Error(w, "duplicate domain", http.StatusConflict)
			return
		}
	}

	if userAgent := r.Header.Get("User-Agent"); !strings.Contains(userAgent, "workos-go/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(
		Organization{
			ID:   "organization_id",
			Name: "Foo Corp",
			Domains: []OrganizationDomain{
				OrganizationDomain{
					ID:     "organization_domain_id",
					Domain: "foo-corp.com",
				},
			},
		})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func TestGenerateLink(t *testing.T) {
	tests := []struct {
		scenario string
		client   *Client
		options  GenerateLinkOpts
		expected string
		err      bool
	}{
		{
			scenario: "Request without API Key returns an error",
			client:   &Client{},
			err:      true,
		},
		{
			scenario: "Request returns link with sso intent",
			client: &Client{
				APIKey: "test",
			},
			options: GenerateLinkOpts{
				Intent:       SSO,
				Organization: "organization_id",
				ReturnURL:    "https://foo-corp.app.com/settings",
			},
			expected: "https://id.workos.test/portal/launch?secret=1234",
		},
		{
			scenario: "Request returns link with dsync intent",
			client: &Client{
				APIKey: "test",
			},
			options: GenerateLinkOpts{
				Intent:       DSync,
				Organization: "organization_id",
				ReturnURL:    "https://foo-corp.app.com/settings",
			},
			expected: "https://id.workos.test/portal/launch?secret=1234",
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(generateLinkTestHandler))
			defer server.Close()

			client := test.client
			client.Endpoint = server.URL
			client.HTTPClient = server.Client()

			link, err := client.GenerateLink(context.Background(), test.options)
			if test.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, link)
		})
	}
}

func generateLinkTestHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "Bearer test" {
		http.Error(w, "bad auth", http.StatusUnauthorized)
		return
	}

	body, err := json.Marshal(struct {
		generateLinkResponse
	}{generateLinkResponse{Link: "https://id.workos.test/portal/launch?secret=1234"}})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
