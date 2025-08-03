package fontl

type Metadata struct {
	Name          string   `json:"name"`
	Source        string   `json:"source"`
	CommercialUse bool     `json:"commercial_use"`
	Projects      []string `json:"projects"`
	Tags          []string `json:"tags"`
}
