package metrics

type Metrics struct {
	caches    int
	hits      int
	misses    int
	uploads   int
	downloads int
}

func (m *Metrics) Init(caches int) {
	m.caches = caches
	m.hits = 0
	m.misses = 0
	m.uploads = 0
	m.downloads = 0
}

func (m *Metrics) Hit() {
	m.hits++
}

func (m *Metrics) Miss() {
	m.misses++
}

func (m *Metrics) Up() {
	m.uploads++
}

func (m *Metrics) Down() {
	m.downloads++
}
