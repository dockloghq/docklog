package main

import "testing"

func TestIsDockLogSelfContainer(t *testing.T) {
	if !isDockLogSelfContainer("docklog", "nginx:latest") {
		t.Fatal("expected docklog name match")
	}
	if !isDockLogSelfContainer("/docklog", "") {
		t.Fatal("expected trimmed docklog name match")
	}
	if !isDockLogSelfContainer("api", "aimldev/docklog:latest") {
		t.Fatal("expected docklog image match")
	}
	if isDockLogSelfContainer("api", "nginx:latest") {
		t.Fatal("expected non-docklog container to be false")
	}
}

func TestIsExcludedContainer(t *testing.T) {
	excludedContainerNames = []string{"redis", "proxy"}

	if !isExcludedContainer("docklog", "nginx:latest") {
		t.Fatal("docklog self must always be excluded")
	}
	if !isExcludedContainer("redis", "redis:7") {
		t.Fatal("expected redis in exclude list")
	}
	if !isExcludedContainer("proxy", "nginx:latest") {
		t.Fatal("expected proxy in exclude list")
	}
	if isExcludedContainer("api", "node:20") {
		t.Fatal("api should not be excluded")
	}
}
