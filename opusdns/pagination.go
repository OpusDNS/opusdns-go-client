package opusdns

func cloneOptions[T any](opts *T) *T {
	pageOpts := new(T)
	if opts != nil {
		*pageOpts = *opts
	}
	return pageOpts
}
