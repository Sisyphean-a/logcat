package app

func cloneFilterState(state FilterState) FilterState {
	cloned := state
	cloned.Saved = append([]SavedFilter(nil), state.Saved...)
	cloned.History = append([]string(nil), state.History...)
	return cloned
}
