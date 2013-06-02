package main

import "testing"

func TestVersion(t *testing.T) {
    if BUILDTIMESTAMP <= 1 {
        t.Fatalf("BUILDTIMESTAMP=%v, want > 1", BUILDTIMESTAMP)
    }

    lenBuildUser := len(BUILDUSER)
    if lenBuildUser <= 0 {
        t.Fatalf("len(BUILDUSER)=%v, want > 0; BUILDUSER=%s", lenBuildUser, BUILDUSER)
    }

    lenBuildHost := len(BUILDHOST)
    if lenBuildHost <= 0 {
        t.Fatalf("len(BUILDHOST)=%v, want > 0; BUILDHOST=%s", lenBuildUser, BUILDHOST)
    }
}

