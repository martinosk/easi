package domain

type Upcaster interface {
	EventType() string
	Upcast(data map[string]interface{}) map[string]interface{}
}

type UpcasterChain []Upcaster

func (chain UpcasterChain) Upcast(eventType string, data map[string]interface{}) map[string]interface{} {
	for _, upcaster := range chain {
		if upcaster.EventType() == eventType {
			data = upcaster.Upcast(data)
		}
	}
	return data
}
