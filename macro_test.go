package spf

import (
	"net"
	"testing"
)

const (
	domain = "matching.com"
	sender = "sender@domain.com"
)

var (
	ip4 = net.IP{10, 11, 12, 13}
	tkn = &token{mechanism: tExp, qualifier: qMinus, value: ""}
)

type MacroTest struct {
	Input  string
	Output string
}

func TestMacroIteration(t *testing.T) {
	testCases := []*MacroTest{
		{"matching.com", "matching.com"},
		{"%%matching.com", "%matching.com"},
		{"%%matching%_%%.com", "%matching %.com"},
		{"matching%-.com", "matching%20.com"},
		{"%%%%%_%-", "%% %20"},
		{"Please email to %{s} end",
			"Please email to sender@domain.com end"},
		{"Please email to %{l} end",
			"Please email to sender end"},
		{"Please email to %{o} end",
			"Please email to domain.com end"},
		{"domain %{d} end",
			"domain matching.com end"},
		{"Address IP %{i} end",
			"Address IP 10.11.12.13 end"},
		{"Address IP %{i1} end",
			"Address IP 13 end"},
		{"Address IP %{i100} end",
			"Address IP 10.11.12.13 end"},
		{"Address IP %{ir} end",
			"Address IP 13.12.11.10 end"},
		{"Address IP %{i2r} end",
			"Address IP 11.10 end"},
		{"Address IP %{i500r} end",
			"Address IP 13.12.11.10 end"},
	}

	parser := newParser(WithResolver(testResolver)).with(stub, sender, domain, ip4)

	for _, test := range testCases {
		tkn.value = test.Input
		result, err := parseMacroToken(parser, tkn)
		if err != nil {
			t.Errorf("Macro %s evaluation failed due to returned error: %v\n",
				test.Input, err)
		}
		if result != test.Output {
			t.Errorf("Macro '%s', evaluation failed, got: '%s',\nexpected '%s'\n",
				test.Input, result, test.Output)
		}
	}
}

// TestMacroExpansionRFCExamples will execute examples from RFC 7208, section
// 7.4
func TestMacroExpansionRFCExamples(t *testing.T) {
	testCases := []*MacroTest{
		{"", ""},
		{"%{s}", "strong-bad@email.example.com"},
		{"%{o}", "email.example.com"},
		{"%{d}", "email.example.com"},
		{"%{d4}", "email.example.com"},
		{"%{d3}", "email.example.com"},
		{"%{d2}", "example.com"},
		{"%{d1}", "com"},
		{"%{dr}", "com.example.email"},
		{"%{d2r}", "example.email"},
		{"%{l}", "strong-bad"},
		{"%{l-}", "strong.bad"},
		{"%{lr}", "strong-bad"},
		{"%{lr-}", "bad.strong"},
		{"%{l1r-}", "strong"},
		{"%{ir}.%{v}._spf.%{d2}",
			"3.2.0.192.in-addr._spf.example.com"},
		{"%{lr-}.lp._spf.%{d2}", "bad.strong.lp._spf.example.com"},
		{"%{lr-}.lp.%{ir}.%{v}._spf.%{d2}",
			"bad.strong.lp.3.2.0.192.in-addr._spf.example.com"},
		{"%{ir}.%{v}.%{l1r-}.lp._spf.%{d2}",
			"3.2.0.192.in-addr.strong.lp._spf.example.com"},
		{"%{d2}.trusted-domains.example.net",
			"example.com.trusted-domains.example.net"},
	}

	parser := newParser(WithResolver(testResolver)).
		with(stub, "strong-bad@email.example.com", "email.example.com", net.IP{192, 0, 2, 3})

	for _, test := range testCases {

		tkn.value = test.Input
		result, err := parseMacroToken(parser, tkn)
		if err != nil {
			t.Errorf("Macro %s evaluation failed due to returned error: %v\n",
				test.Input, err)
		}
		if result != test.Output {
			t.Errorf("Macro '%s', evaluation failed, got: '%s',\nexpected '%s'\n",
				test.Input, result, test.Output)
		}
	}
}

// TODO(zaccone): Fill epected error messages and compare with those returned.
func TestParsingErrors(t *testing.T) {
	testcases := []*MacroTest{
		{"%", ""},
		{"%{?", ""},
		{"%}", ""},
		{"%a", ""},
		{"%", ""},
		{"%{}", ""},
		{"%{", ""},
		{"%{234", ""},
		{"%{2a3}", ""},
		{"%{i2", ""},
		{"%{s2a3}", ""},
		{"%{s2i3}", ""},
		{"%{s2ir-3}", ""},
		{"%{l2a3}", ""},
		{"%{i2a3}", ""},
		{"%{o2a3}", ""},
		{"%{d2a3}", ""},
		{"%{i-2}", ""},
	}

	parser := newParser(WithResolver(testResolver)).with(stub, sender, domain, ip4)

	for _, test := range testcases {

		tkn.value = test.Input
		result, err := parseMacroToken(parser, tkn)

		if result != "" {
			t.Errorf("For input '%s' expected empty result, got '%s' instead\n",
				test.Input, result)
		}

		if err == nil {
			t.Errorf("For input '%s', expected non-empty err, got nil instead and result '%s'\n",
				test.Input, result)
		}
	}
}
