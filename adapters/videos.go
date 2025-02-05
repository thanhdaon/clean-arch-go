package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type VideoService struct {
	client *http.Client
}

func NewVideoService(httpclient *http.Client) VideoService {
	return VideoService{client: httpclient}
}

func (s VideoService) GetAll(ctx context.Context) error {
	var videos []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8001/", nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status is not ok")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &videos); err != nil {
		return err
	}

	for _, video := range videos {
		fmt.Printf("ID: %d, Title: %s \n", video.ID, video.Title)
	}

	return nil
}
