package backend

// TerminalBundleFingerprints returns the compile-time development bundle
// contract. A copy prevents callers changing backend verification policy.
func TerminalBundleFingerprints() map[string]string {
	out := make(map[string]string, len(terminalBundleHashes))
	for name, hash := range terminalBundleHashes {
		out[name] = hash
	}
	return out
}
