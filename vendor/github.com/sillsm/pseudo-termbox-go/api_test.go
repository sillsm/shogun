package termbox

import (
	"testing"
)

type TestReader struct {
	Data []byte
}

// Reconsider. Read should block until there is data available!
func (tr *TestReader) Read(b []byte) (n int, err error) {
	if len(b) < 1 {
		return 666, nil
	}
	//4  and 20. Buffer to be read into smaller than data available.
	if len(b) < len(tr.Data) {
		//fmt.Printf("CA1, Buffer size: %v, Data %v", len(b), tr.Data)
		copy(b, tr.Data)
		tr.Data = tr.Data[len(b):]
		return len(b), nil
	}
	if len(tr.Data) <= len(b) {
		//fmt.Printf("CA1, Buffer size: %v, Data %v", len(b), tr.Data)
		n := len(tr.Data)
		n = n
		copy(b, tr.Data)
		tr.Data = []byte{}
		return n, nil

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
		{[]byte("\x1b[<33;50;13M"), EventMouse},
	}
	for _, entry := range tt {
		reader1.PushData(entry.in)
		ev := tbox.PollEvent()
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
		ev := tbox.PollEvent()
		if ev.Type != entry.evType {
			t.Errorf("On input: %q got %v, want %v", string(entry.in), ev.Type, entry.evType)
		}
	}
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
		in     []byte
		key    Key
		motion bool
		x      int
		y      int
	}{
		// xterm 1006 extended mode documented here: https://www.xfree86.org/4.8.0/ctlseqs.html
		{[]byte("\x1b[<0;49;13M"), MouseLeft, false, 48, 12},
		{[]byte("\x1b[<32;49;13M"), MouseLeft, true, 48, 12}, //Adding 32 sets motion flag.
		{[]byte("\x1b[<80;49;13M"), MouseWheelUp, false, 48, 12},
		{[]byte("\x1b[<112;49;13M"), MouseWheelUp, true, 48, 12},
		{[]byte("\x1b[<21;49;13M"), MouseMiddle, false, 48, 12},
		{[]byte("\x1b[<53;49;13M"), MouseMiddle, true, 48, 12},
		{[]byte("\x1b[<81;49;13M"), MouseWheelDown, false, 48, 12},
		{[]byte("\x1b[<113;49;13M"), MouseWheelDown, true, 48, 12},
		{[]byte("\x1b[<02;1;1M"), MouseRight, false, 0, 0},
		{[]byte("\x1b[<34;1;1M"), MouseRight, true, 0, 0},
		{[]byte("\033[<0;49;13m"), MouseRelease, false, 48, 12}, // xterm 1006 uses (m) for mouse release
		{[]byte("\033[<32;49;13m"), MouseRelease, true, 48, 12}, // independent of motion bit
		{[]byte("\x1b[<23;1;1M"), MouseRelease, false, 0, 0},    // or just regular code
		// x10 sequence
		{[]byte("\033[M\x20\x2e\x2d"), MouseLeft, false, 13, 12},
		{[]byte("\033[M\x40\x2e\x2d"), MouseLeft, true, 13, 12}, // This
		{[]byte("\033[M\x30\x2e\x2d"), MouseLeft, false, 13, 12},
		{[]byte("\033[M\x50\x2e\x2d"), MouseLeft, true, 13, 12},
		{[]byte("\033[M\x60\x2e\x2d"), MouseWheelUp, false, 13, 12},
		{[]byte("\033[M\x70\x2e\x2d"), MouseWheelUp, false, 13, 12},
		{[]byte("\033[M\x80\x2e\x2d"), MouseWheelUp, true, 13, 12},
		{[]byte("\033[M\x90\x2e\x2d"), MouseWheelUp, true, 13, 12},
		{[]byte("\033[M\x21\x2e\x2d"), MouseMiddle, false, 13, 12},
		{[]byte("\033[M\x31\x2e\x2d"), MouseMiddle, false, 13, 12},
		{[]byte("\033[M\x41\x2e\x2d"), MouseMiddle, true, 13, 12},
		{[]byte("\033[M\x51\x2e\x2d"), MouseMiddle, true, 13, 12},
		{[]byte("\033[M\x61\x32\x4d"), MouseWheelDown, false, 17, 44},
		{[]byte("\033[M\x71\x32\x4d"), MouseWheelDown, false, 17, 44},
		{[]byte("\033[M\x81\x32\x4d"), MouseWheelDown, true, 17, 44},
		{[]byte("\033[M\x91\x32\x4d"), MouseWheelDown, true, 17, 44},
		{[]byte("\033[M\x22\x2e\x2d"), MouseRight, false, 13, 12},
		{[]byte("\033[M\x32\x2e\x2d"), MouseRight, false, 13, 12},
		{[]byte("\033[M\x42\x2e\x2d"), MouseRight, true, 13, 12},
		{[]byte("\033[M\x52\x2e\x2d"), MouseRight, true, 13, 12},
		{[]byte("\033[M\x23\x2e\x2d"), MouseRelease, false, 13, 12},
		{[]byte("\033[M\x33\x2e\x2d"), MouseRelease, false, 13, 12},
		{[]byte("\033[M\x43\x2e\x2d"), MouseRelease, true, 13, 12},
		{[]byte("\033[M\x53\x2e\x2d"), MouseRelease, true, 13, 12},
		// urxvt
		{[]byte("\x1b[20;49;13M"), MouseWheelUp, true, 48, 12},
		{[]byte("\x1b[244;49;13M"), MouseWheelUp, false, 48, 12},
		{[]byte("\x1b[21;49;13M"), MouseWheelDown, true, 48, 12},
		{[]byte("\x1b[245;49;13M"), MouseWheelDown, false, 48, 12},
		{[]byte("\x1b[32;49;13M"), MouseLeft, false, 48, 12},
		{[]byte("\x1b[64;100;2M"), MouseLeft, true, 99, 1},
		{[]byte("\x1b[10;49;13M"), MouseRight, true, 48, 12},
		{[]byte("\x1b[42;49;13M"), MouseRight, false, 48, 12},
		{[]byte("\x1b[11;49;13M"), MouseRelease, true, 48, 12},
		{[]byte("\x1b[43;49;13M"), MouseRelease, false, 48, 12},
		{[]byte("\x1b[33;49;13M"), MouseMiddle, false, 48, 12},
		{[]byte("\x1b[65;49;13M"), MouseMiddle, true, 48, 12},
	}
	for _, entry := range tt {
		reader.PushData(entry.in)
		ev := tbox.PollEvent()
		if ev.Key != entry.key {
			t.Errorf("On input: %q, ev.Key != entry.key. got %v, want %v", entry.in, ev.Key, entry.key)
		}
		if ev.MouseX != entry.x {
			t.Errorf("On input: %v, ev.MouseX != entry.x. got %v, want %v", entry.in, ev.MouseX, entry.x)
		}
		if ev.MouseY != entry.y {
			t.Errorf("On input: %v, ev.MouseY != entry.y. got %v, want %v", entry.in, ev.MouseY, entry.y)
		}
		if v := (ev.Mod & ModMotion) != 0; v != entry.motion {
			t.Errorf("On input: %q, motion detection error. got %v, want %v", string(entry.in), v, entry.motion)
		}
	}
}

// Test that unread input buffering still returns correct event processing.
func TestPollEvent2_Buffering(t *testing.T) {
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

	var tt = []struct {
		in  []byte
		key Key
		ch  rune
	}{
		{[]byte("l"), 0, 'l'},
		{[]byte("q"), 0, 'q'},
		{[]byte("z"), 0, 'z'},
		{[]byte("\x1b[<0;49;13M"), MouseLeft, 0},
		{[]byte("\x1b[<32;49;13M"), MouseLeft, 0},
		{[]byte("\x1b[<32;50;13M"), MouseLeft, 0},
		{[]byte("\x1b[<32;51;13M"), MouseLeft, 0},

		{[]byte("\033[<0;49;13m"), MouseRelease, 0}, // xterm 1006 uses (m) for mouse release
		{[]byte("\033[<32;49;13m"), MouseRelease, 0},

		{[]byte("a"), 0, 'a'},
		{[]byte("\x1b[21;49;13M"), MouseWheelDown, 0},
		{[]byte("x"), 0, 'x'},
	}
	// Load up the buffer with the test cases
	for _, entry := range tt {
		reader.PushData(entry.in)
	}

	for _, entry := range tt {
		ev := tbox.PollEvent()
		if ev.Key != entry.key {
			t.Errorf("On input: %q, ev.Key != entry.key. got %v, want %v", entry.in, ev.Key, entry.key)
		}
		if ev.Ch != entry.ch {
			t.Errorf("On input: %v, ev.Ch != entry.ch. got %v, want %v", entry.in, ev.Ch, entry.ch)
		}
	}
}
