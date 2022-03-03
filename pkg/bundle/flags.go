package bundle

// BindFlags define the flags used to generate the bundle report
type BindFlags struct {
	IndexImage       string `json:"image"`
	Filter           string `json:"filter"`
	DisableScorecard bool   `json:"disableScorecard"`
	ContainerEngine  string `json:"containerEngine"`
}
