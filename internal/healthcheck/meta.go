package healthcheck

import "strconv"

type Meta struct {
	id          int
	names       int
	image       int
	status      int
	healthCheck int
}

func (m *Meta) ToString() *metaString {
	return &metaString{
		id:          strconv.Itoa(m.id),
		names:       strconv.Itoa(m.names),
		image:       strconv.Itoa(m.image),
		status:      strconv.Itoa(m.status),
		healthCheck: strconv.Itoa(m.healthCheck),
	}
}

type metaString struct {
	id          string
	names       string
	image       string
	status      string
	healthCheck string
}

func (m *metaString) ToInteger() *Meta {
	id, _ := strconv.Atoi(m.id)
	names, _ := strconv.Atoi(m.names)
	image, _ := strconv.Atoi(m.image)
	status, _ := strconv.Atoi(m.status)
	healthCheck, _ := strconv.Atoi(m.healthCheck)

	return &Meta{
		id:          id,
		names:       names,
		image:       image,
		status:      status,
		healthCheck: healthCheck,
	}
}
