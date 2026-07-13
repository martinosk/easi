package events

type TargetPeriodData struct {
	Year    int `json:"year"`
	Quarter int `json:"quarter"`
}

func targetPeriodEventData(tp *TargetPeriodData) map[string]interface{} {
	if tp == nil {
		return nil
	}
	return map[string]interface{}{"year": tp.Year, "quarter": tp.Quarter}
}
