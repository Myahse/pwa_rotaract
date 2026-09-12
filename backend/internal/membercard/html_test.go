package membercard

import (
	"strings"
	"testing"
	"time"
)

func TestRenderHTML(t *testing.T) {
	html, err := RenderHTML(CardData{
		FullName:    "Aya Kouassi",
		ClubName:    "Rotaract Abidjan Plateau",
		CardNumber:  "RCI-ABIDJ-0042",
		MemberSince: "09/2024",
		IssuedAt:    time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"AYA KOUASSI", "Rotaract Abidjan Plateau", "RCI · ABIDJ · 0042", "Émise le 12/09/2026"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected html to contain %q", want)
		}
	}
}
