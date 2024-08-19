package fasttime

import (
	"testing"
	"time"
)

func TestUnixTimestamp(t *testing.T) {
	tsExpected := time.Now().Unix()
	ts := UnixTimestamp()
	if ts-tsExpected > 1 {
		t.Fatalf("unexpected UnixTimestamp; got %d; want %d", ts, tsExpected)
	}
	time.Sleep(time.Second)
	diff := time.Since(Time())
	if diff > time.Millisecond*21 {
		t.Errorf("time is not correct %v", diff)
	}
}

func TestUnixDate(t *testing.T) {
	dateExpected := time.Now().Unix() / (24 * 3600)
	date := UnixDate()
	if date-dateExpected > 1 {
		t.Fatalf("unexpected UnixDate; got %d; want %d", date, dateExpected)
	}
}

func TestUnixHour(t *testing.T) {
	hourExpected := time.Now().Unix() / 3600
	hour := UnixHour()
	if hour-hourExpected > 1 {
		t.Fatalf("unexpected UnixHour; got %d; want %d", hour, hourExpected)
	}
}

func TestUnixMinute(t *testing.T) {
	minuteExpected := time.Now().Unix() / 60
	minute := UnixMinute()
	if minute-minuteExpected > 1 {
		t.Fatalf("unexpected UnixMinute; got %d; want %d", minute, minuteExpected)
	}
}

func TestTime(t *testing.T) {
	tmExpected := time.Unix(int64(UnixTimestamp()), 0)
	tm := Time()
	if tm.Sub(tmExpected) > time.Second {
		t.Fatalf("unexpected Time; got %s; want %s", tm, tmExpected)
	}
}
