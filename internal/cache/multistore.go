package cache

import "time"

type MultiStore struct {
	Primary   Store
	Secondary Store // optional
}

func (m *MultiStore) GetTaskView(key string) (*TaskView, bool, error) {
	if m.Primary != nil {
		if v, ok, err := m.Primary.GetTaskView(key); err == nil && ok {
			return v, ok, nil
		}
	}
	if m.Secondary != nil {
		return m.Secondary.GetTaskView(key)
	}
	return nil, false, nil
}

func (m *MultiStore) SetTaskView(key string, view *TaskView, ttl time.Duration) error {
	if m.Primary != nil {
		_ = m.Primary.SetTaskView(key, view, ttl)
	}
	if m.Secondary != nil {
		_ = m.Secondary.SetTaskView(key, view, ttl)
	}
	return nil
}

func (m *MultiStore) SetNull(key string, ttl time.Duration) error {
	if m.Primary != nil {
		_ = m.Primary.SetNull(key, ttl)
	}
	if m.Secondary != nil {
		_ = m.Secondary.SetNull(key, ttl)
	}
	return nil
}

func (m *MultiStore) Delete(key string) error {
	if m.Primary != nil {
		_ = m.Primary.Delete(key)
	}
	if m.Secondary != nil {
		_ = m.Secondary.Delete(key)
	}
	return nil
}
