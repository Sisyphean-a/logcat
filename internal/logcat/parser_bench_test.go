package logcat

import "testing"

func BenchmarkParseThreadtimeLine(b *testing.B) {
	line := `06-04 16:42:18.479 10665 10665 I chromium: [INFO:CONSOLE(618)] "[H5] connected", source: http://127.0.0.1/app.js (618)`

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ParseThreadtimeLine("device-1", line); err != nil {
			b.Fatal(err)
		}
	}
}
