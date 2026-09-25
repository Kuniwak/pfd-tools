package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

const dupPFD = `<mxfile><diagram id="g0" name="P0"><mxGraphModel><root>` +
	`<mxCell id="0"/><mxCell id="1" parent="0"/>` +
	`<mxCell id="p1" value="P1: 作る" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="0" y="0" width="120" height="80" as="geometry"/></mxCell>` +
	`<mxCell id="d1a" value="D1: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="200" y="0" width="120" height="80" as="geometry"/></mxCell>` +
	`<mxCell id="d1b" value="D1: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="200" y="200" width="120" height="80" as="geometry"/></mxCell>` +
	`<mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d1b" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>` +
	`</root></mxGraphModel></diagram></mxfile>`

func TestMainCommandByArgs_StdinMarksDuplicates(t *testing.T) {
	spy := cli.SpyProcInout(dupPFD)
	exit := MainCommandByArgs(nil, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}

	got := spy.Stdout.String()
	if !strings.Contains(got, `id="d1a" value="D1＊: 出力"`) {
		t.Errorf("複製の d1a にマークが付いていない:\n%s", got)
	}
	if !strings.Contains(got, `id="d1b" value="D1: 出力"`) {
		t.Errorf("正本の d1b が無印でない:\n%s", got)
	}
}

func TestMainCommandByArgs_Inplace(t *testing.T) {
	tmpPath := filepath.Join(t.TempDir(), "pfd.drawio")
	if err := os.WriteFile(tmpPath, []byte(dupPFD), 0644); err != nil {
		t.Fatal(err)
	}
	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"-inplace", tmpPath}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}
	if spy.Stdout.Len() != 0 {
		t.Errorf("stdout should be empty in -inplace mode, got %q", spy.Stdout.String())
	}
	out, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte(`value="D1＊: 出力"`)) {
		t.Errorf("file should be rewritten in place:\n%s", out)
	}
}

func TestMainCommandByArgs_OnlyPageLeavesOtherPagesUntouched(t *testing.T) {
	spy := cli.SpyProcInout(dupPFD)
	exit := MainCommandByArgs([]string{"-only-page", "P7"}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}
	if got := spy.Stdout.String(); strings.Contains(got, "＊") {
		t.Errorf("-only-page の対象外のページにマークが付いた:\n%s", got)
	}
}

func TestMainCommandByArgs_PNGInput(t *testing.T) {
	xmlPath := filepath.Join(t.TempDir(), "pfd.drawio")
	if err := os.WriteFile(xmlPath, []byte(dupPFD), 0644); err != nil {
		t.Fatal(err)
	}
	pngPath := pngtest.WriteTempPNG(t, xmlPath)

	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{pngPath}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}

	got := pngtest.MxfileInPNG(t, spy.Stdout.Bytes())
	if !strings.Contains(got, `value="D1＊: 出力"`) {
		t.Errorf("PNG に埋めた XML にマークが付いていない:\n%s", got)
	}
}

func TestMainCommandByArgs_ShortHelp(t *testing.T) {
	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"--short-help"}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0", exit)
	}
	if got := spy.Stdout.String(); got != ShortHelp+"\n" {
		t.Errorf("stdout = %q, want %q", got, ShortHelp+"\n")
	}
}
