package chartmuseum

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

type (
	// ChartInfo holds a chart's name, version as well as optional org and repo attributes.
	ChartInfo struct {
		// Name of the Chart
		Name *string
		// Version of the Chart
		Version *string
		// Org the Chart belongs to
		Org *string
		// Repo the Chart belongs to
		Repo *string
	}
)

func (c ChartInfo) String() string {
	s := fmt.Sprintf("%s-%s", *c.Name, *c.Version)
	if *c.Org != "" {
		return fmt.Sprintf("%s/%s/%s", *c.Org, *c.Repo, s)
	}
	if *c.Repo != "" {
		return fmt.Sprintf("%s/%s", *c.Repo, s)
	}
	return s
}

// ChartService handles communication with the Chart Manipulation
// related methods of the ChartMuseum API.
type ChartService service

// UploadChart uploads a Helm chart to a ChartMuseum server
func (s *ChartService) UploadChart(ctx context.Context, c *ChartInfo, file *os.File) (*Response, error) {
	u := "api/charts"
	if *c.Org != "" {
		if *c.Repo == "" {
			return nil, errors.Errorf("Repo required if Org is provided")
		}
		u = fmt.Sprintf("api/%s/%s/charts", *c.Org, *c.Repo)
	} else if *c.Repo != "" {
		u = fmt.Sprintf("api/%s/charts", *c.Repo)
	}
	return s.uploadChartHelper(ctx, u, file)
}

// DeleteChart deletes a Helm chart from a ChartMuseum server
func (s *ChartService) DeleteChart(ctx context.Context, c *ChartInfo) (*Response, error) {
	u := fmt.Sprintf("api/charts/%s/%s", *c.Name, *c.Version)
	if *c.Org != "" {
		if *c.Repo == "" {
			return nil, errors.Errorf("Repo required if Org is provided")
		}
		u = fmt.Sprintf("api/%s/%s/charts/%s/%s", *c.Org, *c.Repo, *c.Name, *c.Version)
	} else if *c.Repo != "" {
		u = fmt.Sprintf("api/%s/charts/%s/%s", *c.Repo, *c.Name, *c.Version)
	}
	return s.deleteChartHelper(ctx, u)
}

// deleteChartHelper prepares and executes the upload request
func (s *ChartService) uploadChartHelper(ctx context.Context, u string, file *os.File) (*Response, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to access file")
	}
	if stat.IsDir() {
		return nil, errors.New("Chart to upload can't be a directory")
	}
	//mediaType := mime.TypeByExtension(filepath.Ext(file.Name()))
	mediaType, _ := detectContentType(file)
	req, err := s.client.NewUploadRequest(u, file, stat.Size(), mediaType)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating upload request")
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to do upload request")
	}
	return resp, nil
}

// detectContentType returns a valid content-type and "application/octet-stream" if error or no match
func detectContentType(file *os.File) (string, error) {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "application/octet-stream", err
	}

	// Reset the read pointer.
	file.Seek(0, 0)

	// Always returns a valid content-type and "application/octet-stream" if no others seemed to match.
	return http.DetectContentType(buffer), nil
}

// deleteChartHelper prepares and executes the delete request
func (s *ChartService) deleteChartHelper(ctx context.Context, u string) (*Response, error) {
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating delete request")
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to do delete request")
	}
	return resp, nil
}
