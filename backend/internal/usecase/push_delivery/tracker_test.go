package push_delivery_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/auth"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/push_delivery"
)

func newTracker() (*push_delivery.Tracker, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))
	return push_delivery.NewTracker(auth.NewPushAckTokenIssuer("secret"), logger), buf
}

func TestAckShownLogsInfo(t *testing.T) {
	tr, buf := newTracker()
	now := time.Now()
	token, _ := tr.Track("u1", "fcm.googleapis.com", now)
	if !tr.Ack(push_delivery.Report{Token: token, Status: push_delivery.StatusShown, Permission: "granted"}, now) {
		t.Fatal("expected valid token")
	}
	if !strings.Contains(buf.String(), "displayed on client") {
		t.Fatalf("unexpected log: %s", buf)
	}
}

func TestAckDeniedLogsError(t *testing.T) {
	tr, buf := newTracker()
	now := time.Now()
	token, _ := tr.Track("u1", "fcm.googleapis.com", now)
	tr.Ack(push_delivery.Report{Token: token, Status: push_delivery.StatusFailed, Permission: "denied"}, now)
	out := buf.String()
	if !strings.Contains(out, "level=ERROR") || !strings.Contains(out, "permission is denied") {
		t.Fatalf("unexpected log: %s", out)
	}
}

func TestAckRejectsInvalidToken(t *testing.T) {
	tr, _ := newTracker()
	if tr.Ack(push_delivery.Report{Token: "bogus"}, time.Now()) {
		t.Fatal("expected invalid token to be rejected")
	}
}

func TestSweepWarnsOnceWhenUnacked(t *testing.T) {
	tr, buf := newTracker()
	now := time.Now()
	tr.Track("u1", "fcm.googleapis.com", now)
	tr.Sweep(now.Add(push_delivery.AckTimeout - time.Second))
	if buf.Len() != 0 {
		t.Fatalf("no warning expected yet: %s", buf)
	}
	tr.Sweep(now.Add(push_delivery.AckTimeout + time.Second))
	tr.Sweep(now.Add(push_delivery.AckTimeout + 2*time.Second))
	if n := strings.Count(buf.String(), "not acknowledged"); n != 1 {
		t.Fatalf("expected exactly one warning, got %d: %s", n, buf)
	}
}

func TestForgetSuppressesWarning(t *testing.T) {
	tr, buf := newTracker()
	now := time.Now()
	token, _ := tr.Track("u1", "fcm.googleapis.com", now)
	tr.Forget(token)
	tr.Sweep(now.Add(push_delivery.AckTimeout + time.Second))
	if buf.Len() != 0 {
		t.Fatalf("no warning expected: %s", buf)
	}
}
