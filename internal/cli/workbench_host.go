package cli

// The optional native endpoint lives only for this workbench invocation.
type workbenchHost struct {
	focus func() bool
	close func()
}

func (h *workbenchHost) Close() {
	if h != nil && h.close != nil {
		h.close()
	}
}
