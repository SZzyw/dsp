package config

import "testing"

func TestStoreCurrentInputFileAccessors(t *testing.T) {
	store := &Store{cfg: Config{}}
	if !store.CurrentInputFileFlashEnabled() || !store.CurrentInputFileProEnabled() || !store.CurrentInputFileVisionEnabled() {
		t.Fatal("expected current input file enabled for all model families by default")
	}
	for _, model := range []string{"deepseek-v4-flash", "deepseek-v4-pro", "deepseek-v4-vision"} {
		if !store.CurrentInputFileEnabledForModel(model) {
			t.Fatalf("expected current input file enabled for model %s by default", model)
		}
	}

	flash := false
	pro := true
	vision := false
	store.cfg.CurrentInputFile = CurrentInputFileConfig{Flash: &flash, Pro: &pro, Vision: &vision}
	if store.CurrentInputFileFlashEnabled() {
		t.Fatal("expected flash current input file disabled")
	}
	if !store.CurrentInputFileProEnabled() {
		t.Fatal("expected pro current input file enabled")
	}
	if store.CurrentInputFileVisionEnabled() {
		t.Fatal("expected vision current input file disabled")
	}
	if store.CurrentInputFileEnabledForModel("deepseek-v4-flash-search") {
		t.Fatal("expected flash-search to inherit flash toggle")
	}
	if !store.CurrentInputFileEnabledForModel("deepseek-v4-pro-nothinking") {
		t.Fatal("expected pro-nothinking to inherit pro toggle")
	}
	if store.CurrentInputFileEnabledForModel("deepseek-v4-vision") {
		t.Fatal("expected vision to inherit disabled toggle")
	}
}
