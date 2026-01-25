package adapter

import (
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/rs/zerolog/log"
)

// DynamicDropboxClient wraps the Dropbox files client with automatic token refresh
type DynamicDropboxClient struct {
	tokenManager *DropboxTokenManager
	client       files.Client
}

// NewDynamicDropboxClient creates a new dynamic Dropbox client
func NewDynamicDropboxClient(tokenManager *DropboxTokenManager) *DynamicDropboxClient {
	// Initial client with current token
	validToken, _ := tokenManager.GetValidToken()
	cfg := dropbox.Config{Token: validToken}

	return &DynamicDropboxClient{
		tokenManager: tokenManager,
		client:       files.New(cfg),
	}
}

// refreshClient refreshes the underlying client with a new valid token
func (dc *DynamicDropboxClient) refreshClient() error {
	validToken, err := dc.tokenManager.GetValidToken()
	if err != nil {
		log.Error().Err(err).Msg("failed to get valid token")
		return err
	}

	cfg := dropbox.Config{Token: validToken}
	dc.client = files.New(cfg)

	return nil
}

// Upload wraps the files.Upload method with automatic token refresh
func (dc *DynamicDropboxClient) Upload(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
	// Ensure we have a valid token
	if err := dc.refreshClient(); err != nil {
		log.Error().Err(err).Msg("failed to refresh client before upload")
		return nil, err
	}

	return dc.client.Upload(arg, content)
}

// UploadSessionStart wraps the UploadSessionStart method with automatic token refresh
func (dc *DynamicDropboxClient) UploadSessionStart(arg *files.UploadSessionStartArg, content io.Reader) (*files.UploadSessionStartResult, error) {
	// Ensure we have a valid token
	if err := dc.refreshClient(); err != nil {
		log.Error().Err(err).Msg("failed to refresh client before upload session start")
		return nil, err
	}

	return dc.client.UploadSessionStart(arg, content)
}

// UploadSessionAppendV2 wraps the UploadSessionAppendV2 method with automatic token refresh
func (dc *DynamicDropboxClient) UploadSessionAppendV2(arg *files.UploadSessionAppendArg, content io.Reader) error {
	// Ensure we have a valid token
	if err := dc.refreshClient(); err != nil {
		log.Error().Err(err).Msg("failed to refresh client before upload session append")
		return err
	}

	return dc.client.UploadSessionAppendV2(arg, content)
}

// UploadSessionFinish wraps the UploadSessionFinish method with automatic token refresh
func (dc *DynamicDropboxClient) UploadSessionFinish(arg *files.UploadSessionFinishArg, content io.Reader) (*files.FileMetadata, error) {
	// Ensure we have a valid token
	if err := dc.refreshClient(); err != nil {
		log.Error().Err(err).Msg("failed to refresh client before upload session finish")
		return nil, err
	}

	return dc.client.UploadSessionFinish(arg, content)
}

// GetUnderlyingClient returns the underlying files.Client (use with caution)
func (dc *DynamicDropboxClient) GetUnderlyingClient() files.Client {
	return dc.client
}
