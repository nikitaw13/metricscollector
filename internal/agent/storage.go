package agent

// Интерфейс для взаимодействия с хранилищем
// Единственный потребитель — тесты
type Storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

// Тип для хранения агентских метрик
type AgentStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

var _ Storage = (*AgentStorage)(nil)

func NewAgentStorage() *AgentStorage {
	return &AgentStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

func (ms *AgentStorage) UpdateGauge(name string, value float64) error {
	ms.gauge[name] = value
	return nil
}

func (ms *AgentStorage) UpdateCounter(name string, value int64) error {
	ms.counter[name] += value
	return nil
}

func (ms *AgentStorage) GetGauge(name string) (value float64, err error) {
	return ms.gauge[name], nil
}

func (ms *AgentStorage) GetCounter(name string) (value int64, err error) {
	return ms.counter[name], nil
}

func (ms *AgentStorage) GetAllGauges() (m map[string]float64) {
	return ms.gauge
}

func (ms *AgentStorage) GetAllCounters() (m map[string]int64) {
	return ms.counter
}
