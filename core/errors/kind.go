package errors

type Kind struct {
	slug string
}

func (r Kind) String() string {
	return r.slug
}

func (r Kind) isZero() bool {
	return r.slug == ""
}

func NewKind(slug string) Kind {
	return Kind{slug: slug}
}
