package cloudfront

import (
	"testing"
)

func TestNewLogEntry(t *testing.T) {
	sampleLine := `2015-12-16   16:07:35        SIN3    181989  112.198.11.111  GET     1234587.cloudfront.net /dir/1.jpg   200     https://www.site.com/index.html Mozilla/5.0%2520(Linux;%2520Android%25204.4.2;%2520MyPhone%2520Rio%25202%2520Craze%2520Build/KOT49H)%2520AppleWebKit/537.36%2520(KHTML,%2520like%2520Gecko)%2520Version/4.0%2520Chrome/30.0.0.0%2520Mobile%2520Safari/537.36        -       -       Hit     kH5h287g285hksg5lajqu0U4jsfh==        sdflkjsdf.cloudfront.net https   561     0.045   -       TLSv1.2 ECDHE-RSA-AES128-GCM-SHA256     Hit`

	logEntry, err := NewLogEntry(sampleLine)
	if err != nil {
		t.Errorf("NewLogEntry returned error: %s\n", err)
	}
	t.Logf("NewLogEntry succeeded: %s\n", logEntry)

	if logEntry.Date != "2015-12-16" {
		t.Error("Error parsing LogEntry.Date")
	} else {
		t.Log("LogEntry.Date parsed successfully")
	}

	if logEntry.Cache != "Hit" {
		t.Error("Error parsing LogEntry.Cache")
	} else {
		t.Log("LogEntry.Cache parsed successfully")
	}
}
