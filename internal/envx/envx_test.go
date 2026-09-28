package envx

import (
	"reflect"
	"testing"
)

func TestParseWSLENV(t *testing.T) {
	got := ParseWSLENV("WT_SESSION:WT_PROFILE_ID:BASH_ENV/u")
	want := []Entry{
		{Name: "WT_SESSION"},
		{Name: "WT_PROFILE_ID"},
		{Name: "BASH_ENV", Flags: "u"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestParseWSLENVEscapes(t *testing.T) {
	if got := ParseWSLENV(""); got != nil {
		t.Fatalf("empty: got %+v", got)
	}
	got := ParseWSLENV("A/p:FOO:BAR/lu")
	want := []Entry{{"A", "p"}, {"FOO", ""}, {"BAR", "lu"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestSuggestRoundtrip(t *testing.T) {
	in := "WT_SESSION:WT_PROFILE_ID:BASH_ENV/u"
	if got := Suggest(ParseWSLENV(in)); got != in {
		t.Fatalf("got %q want %q", got, in)
	}
	if got := Suggest(nil); got != "" {
		t.Fatalf("empty: got %q", got)
	}
}
