package opusdns

import (
	"context"
	"net/url"

	"github.com/opusdns/opusdns-go-client/models"
)

// HostsService provides methods for managing host objects.
type HostsService struct {
	client *Client
}

// CreateHost creates a new host object.
func (s *HostsService) CreateHost(ctx context.Context, req *models.HostCreateRequest) (*models.Host, error) {
	path := s.client.http.BuildPath("hosts")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var host models.Host
	if err := s.client.http.DecodeResponse(resp, &host); err != nil {
		return nil, err
	}

	return &host, nil
}

// GetHost retrieves a host object by either its ID or its hostname.
func (s *HostsService) GetHost(ctx context.Context, reference string) (*models.Host, error) {
	path := s.client.http.BuildPath("hosts", url.PathEscape(reference))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var host models.Host
	if err := s.client.http.DecodeResponse(resp, &host); err != nil {
		return nil, err
	}

	return &host, nil
}

// UpdateHost updates the IP addresses of a host object, referenced by either its ID or
// its hostname.
func (s *HostsService) UpdateHost(ctx context.Context, reference string, req *models.HostUpdateRequest) (*models.Host, error) {
	path := s.client.http.BuildPath("hosts", url.PathEscape(reference))

	resp, err := s.client.http.Put(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var host models.Host
	if err := s.client.http.DecodeResponse(resp, &host); err != nil {
		return nil, err
	}

	return &host, nil
}

// DeleteHost deletes a host object, referenced by either its ID or its hostname. It is
// only possible when the host is not in use.
func (s *HostsService) DeleteHost(ctx context.Context, reference string) error {
	path := s.client.http.BuildPath("hosts", url.PathEscape(reference))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}
