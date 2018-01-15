package termbox

import (
	_ "fmt"
	"testing"
)

type TestReader struct {
	Data []byte
}

func (tr *TestReader) Read(b []byte) (n int, err error) {
	if len(b) < 1 {
		return 0, nil
	}
	//4  and 20
	if len(b) < len(tr.Data) {
		copy(b, tr.Data)
		tr.Data = tr.Data[len(b):]
		return len(b), nil
	}
	if len(tr.Data) <= len(b) {
		copy(b, tr.Data)
		tr.Data = nil
		return len(b), nil

	}
	return 0, nil
}

func (tr *TestReader) PushData(b []byte) {
	tr.Data = append(tr.Data, b...)
}

// Table driven test for random inputs on a reader
func TestPollEvent2_Normal(t *testing.T) {
	// Set up client
	tbox := NewClient()
	err := tbox.Init()
	if err != nil {
		t.Errorf("Error: tbox.Init()")
	}
	defer tbox.Close()

	// Set up inputs
	reader1 := &TestReader{}
	tbox.In = reader1

	// Set up test table
	var tt = []struct {
		in     []byte
		evType EventType
	}{
		{[]byte("\x1b[<32;49;13M"), EventMouse},
		{[]byte("\x1b[<32;49;13M"), EventMouse},
	}
	for _, entry := range tt {
		reader1.PushData(entry.in)
		ev := tbox.PollEvent2()
		if ev.Type != entry.evType {
			//ev.
			t.Errorf("On input: %q got %v, want %v", string(entry.in), ev.Type, entry.evType)
		}
	}

	// Demonstrate you can swap out input readers in the middle of a process.
	reader2 := &TestReader{}
	tbox.In = reader2
	for _, entry := range tt {
		reader2.PushData(entry.in)
		ev := tbox.PollEvent2()
		if ev.Type != entry.evType {
			t.Errorf("On input: %q got %v, want %v", string(entry.in), ev.Type, entry.evType)
		}
	}
	tt = tt

}

// Test non-pathological mouse events.
func TestPollEvent2_MouseEvents(t *testing.T) {
	// Set up client
	tbox := NewClient()
	err := tbox.Init()
	if err != nil {
		t.Errorf("Error: tbox.Init()")
	}
	defer tbox.Close()

	// Set up inputs
	reader := &TestReader{}
	tbox.In = reader

	// Set up test table
	var tt = []struct {
		in  []byte
		key Key
		x   int
		y   int
	}{
		// xterm 1006 extended mode
		{[]byte("\x1b[<0;49;13M"), MouseLeft, 48, 12},
		{[]byte("\x1b[<32;49;13M"), MouseLeft, 48, 12}, //Adding 32 sets motion flag.
		{[]byte("\x1b[<80;49;13M"), MouseWheelUp, 48, 12},
		{[]byte("\x1b[<21;49;13M"), MouseMiddle, 48, 12},
		{[]byte("\x1b[<81;49;13M"), MouseWheelDown, 48, 12},
		{[]byte("\x1b[<02;1;1M"), MouseRight, 0, 0},
		{[]byte("\033[<32;49;13m"), MouseRelease, 48, 12}, // xterm 1006 uses (m) for mouse release
		{[]byte("\x1b[<23;1;1M"), MouseRelease, 0, 0},     // or just regular code
		// x10 sequence, documented here: https://www.xfree86.org/4.8.0/ctlseqs.html
		//{[]byte("\033[M\x00\x2e\x2d"), MouseLeft, 13, 12},
		{[]byte("\033[M\x20\x2e\x2d"), MouseLeft, 13, 12},
		{[]byte("\033[M\x30\x2e\x2d"), MouseLeft, 13, 12},
		{[]byte("\033[M\x80\x2e\x2d"), MouseWheelUp, 13, 12},
		{[]byte("\033[M\x90\x2e\x2d"), MouseWheelUp, 13, 12},
		{[]byte("\033[M\x21\x2e\x2d"), MouseMiddle, 13, 12},
		{[]byte("\033[M\x31\x2e\x2d"), MouseMiddle, 13, 12},
		{[]byte("\033[M\x81\x32\x4d"), MouseWheelDown, 17, 44},
		{[]byte("\033[M\x91\x32\x4d"), MouseWheelDown, 17, 44},
		{[]byte("\033[M\x22\x2e\x2d"), MouseRight, 13, 12},
		{[]byte("\033[M\x32\x2e\x2d"), MouseRight, 13, 12},
		{[]byte("\033[M\x23\x2e\x2d"), MouseRelease, 13, 12},
		{[]byte("\033[M\x33\x2e\x2d"), MouseRelease, 13, 12},
		// urxvt
		{[]byte("\x1b[20;49;13M"), MouseWheelUp, 48, 12},
		{[]byte("\x1b[21;49;13M"), MouseWheelDown, 48, 12},
		{[]byte("\x1b[32;49;13M"), MouseLeft, 48, 12},
		{[]byte("\x1b[10;49;13M"), MouseRight, 48, 12},
		{[]byte("\x1b[42;49;13M"), MouseRight, 48, 12}, //Adding 32 sets motion flag.
		{[]byte("\x1b[43;49;13M"), MouseRelease, 48, 12},
		{[]byte("\x1b[33;49;13M"), MouseMiddle, 48, 12},
	}
	for _, entry := range tt {
		reader.PushData(entry.in)
		ev := tbox.PollEvent2()
		if ev.Key != entry.key {
			t.Errorf("On input: %q, ev.Key != entry.key. got %v, want %v", string(entry.in), ev.Key, entry.key)
		}
		if ev.MouseX != entry.x {
			t.Errorf("On input: %v, ev.MouseX != entry.x. got %v, want %v", entry.in, ev.MouseX, entry.x)
		}
		if ev.MouseY != entry.y {
			t.Errorf("On input: %v, ev.MouseY != entry.y. got %v, want %v", entry.in, ev.MouseY, entry.y)
		}
	}
}

func TestPollEvent2_MouseMotion(t *testing.T) {
	// Set up client
	tbox := NewClient()
	err := tbox.Init()
	if err != nil {
		t.Errorf("Error: tbox.Init()")
	}
	defer tbox.Close()

	// Set up inputs
	reader := &TestReader{}
	tbox.In = reader

	// Set up test table
	var tt = []struct {
		in []byte
		m  Modifier
	}{
		{[]byte("\x1b[<02;1;1M"), ModMotion}, // urxvt
		{[]byte("\x1b[<32;49;13M"), ModNothing},
		{[]byte("\x1b[<80;49;13M"), ModMotion},
		{[]byte("\x1b[<21;49;13M"), ModMotion},
		{[]byte("\x1b[<81;49;13M"), ModMotion},
		{[]byte("\x1b[<02;1;1M"), ModMotion},
		{[]byte("\033[<32;49;13m"), ModNothing}, // xterm 1006 uses (m) for mouse release
		//{[]byte("\x1b[<23;1;1M"), ModNothing},
	}
	for _, entry := range tt {
		reader.PushData(entry.in)
		ev := tbox.PollEvent2()
		if ev.Mod&entry.m != 0 {
			t.Errorf("On input: %q, Mod. got %v, want %v", string(entry.in), ev.Mod, entry.m)
		}
	}
}
