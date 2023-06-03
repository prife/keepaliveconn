package linebreak

import (
	"strings"
)

/*
WARNING: could not locate iTunesMetadata.plist in archive!
WARNING: could not locate Payload/WebDriverAgentRunner-Runner.app/SC_Info/WebDriverAgentRunner-Runner.sinf in archive!
Copying 'resources/wda-v3.12.0.ipa' to device... DONE.
Installing 'com.facebook.WebDriverAgentRunner.xctrunner'
Install: CreatingStagingDirectory (5%)
Install: ExtractingPackage (15%)
Install: InspectingPackage (20%)
Install: TakingInstallLock (20%)
Install: PreflightingApplication (30%)
Install: InstallingEmbeddedProfile (30%)
Install: VerifyingApplication (40%)
ERROR: Install failed. Got error "ApplicationVerificationFailed" with code 0xe8008015: ...
*/

type LineWriter struct {
	buf strings.Builder
	Log func(line string)
}

func checkLineBreak(b []byte) bool {
	for _, c := range b {
		if c == '\n' {
			return true
		}
	}

	return false
}

func (l *LineWriter) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}

	n = len(b)
	err = nil

	l.buf.Write(b)
	lastChar := b[len(b)-1]
	if checkLineBreak(b) {
		lines := strings.Split(l.buf.String(), "\n")
		l.buf.Reset()
		if lastChar != '\n' {
			lastLine := lines[len(lines)-1]
			l.buf.WriteString(lastLine)
		}
		lines = lines[:len(lines)-1]
		for _, line := range lines {
			l.Log(line)
		}
	}

	return
}
func (l *LineWriter) LastLine() string {
	if l.buf.Len() > 0 {
		return l.buf.String()
	}
	return ""
}

func (l *LineWriter) HasLastLine() bool {
	return l.buf.Len() > 0
}
