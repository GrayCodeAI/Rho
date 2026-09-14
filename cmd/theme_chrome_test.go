package cmd

import "testing"

func TestApplyThemeTauEnablesMinimalChrome(t *testing.T) {
	prev := minimalChrome
	t.Cleanup(func() { ApplyTheme("dark"); minimalChrome = prev })

	ApplyTheme("tau")
	if !minimalChrome {
		t.Fatal("ApplyTheme(tau) should enable minimal chrome")
	}
	// Tau keeps only the bottom hairline: no top rule, no side rules.
	if inputBorderStyle.GetBorderTop() || inputBorderStyle.GetBorderLeft() || inputBorderStyle.GetBorderRight() {
		t.Error("tau input border should draw only the bottom hairline")
	}
	if !inputBorderStyle.GetBorderBottom() {
		t.Error("tau input border should keep the bottom hairline")
	}
}

func TestApplyThemeDarkUsesBoxChrome(t *testing.T) {
	prev := minimalChrome
	t.Cleanup(func() { ApplyTheme("dark"); minimalChrome = prev })

	ApplyTheme("dark")
	if minimalChrome {
		t.Fatal("ApplyTheme(dark) should not enable minimal chrome")
	}
	// The default chrome draws a top and bottom rule around the input.
	if !inputBorderStyle.GetBorderTop() || !inputBorderStyle.GetBorderBottom() {
		t.Error("dark input border should keep top and bottom rules")
	}
}

func TestApplyThemeUnknownIsNoop(t *testing.T) {
	ApplyTheme("dark")
	before := minimalChrome
	ApplyTheme("does-not-exist")
	if minimalChrome != before {
		t.Error("ApplyTheme with an unknown name must not change state")
	}
}
