package listener

import (
	"encoding/json"
	"os"
)

type CamEvent struct {
	Before struct {
		ID            string        `json:"id"`
		Camera        string        `json:"camera"`
		FrameTime     float64       `json:"frame_time"`
		Label         string        `json:"label"`
		TopScore      float64       `json:"top_score"`
		FalsePositive bool          `json:"false_positive"`
		StartTime     float64       `json:"start_time"`
		EndTime       interface{}   `json:"end_time"`
		Score         float64       `json:"score"`
		Box           []int         `json:"box"`
		Area          int           `json:"area"`
		Region        []int         `json:"region"`
		CurrentZones  []interface{} `json:"current_zones"`
		EnteredZones  []interface{} `json:"entered_zones"`
		Thumbnail     interface{}   `json:"thumbnail"`
	} `json:"before"`
	After struct {
		ID            string        `json:"id"`
		Camera        string        `json:"camera"`
		FrameTime     float64       `json:"frame_time"`
		Label         string        `json:"label"`
		TopScore      float64       `json:"top_score"`
		FalsePositive bool          `json:"false_positive"`
		StartTime     float64       `json:"start_time"`
		EndTime       interface{}   `json:"end_time"`
		Score         float64       `json:"score"`
		Box           []int         `json:"box"`
		Area          int           `json:"area"`
		Region        []int         `json:"region"`
		CurrentZones  []interface{} `json:"current_zones"`
		EnteredZones  []interface{} `json:"entered_zones"`
		Thumbnail     interface{}   `json:"thumbnail"`
	} `json:"after"`
	Type string `json:"type"`
}

func EventHandle(data []byte) (string, bool) {
	var event CamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		panic(err)
	}
	if isPerson(event) && isCamera(event) {
		return event.After.ID, true
	}

	return "", false
}

func isPerson(event CamEvent) bool {
	return event.After.Label == "person" && event.Type == "new"
}

func isCamera(event CamEvent) bool {
	camName := os.Getenv("CAM_NAME")
	return event.After.Camera == camName
}
