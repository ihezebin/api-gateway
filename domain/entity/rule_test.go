package entity

import "testing"

func TestFindPath(t *testing.T) {

	rule := &Rule{
		Uris: []Uri{
			{
				Paths: []string{
					"/*",
				},
				Endpoint: "test",
				Rewrite: map[string]string{
					"/api": "/",
				},
			},
		},
	}

	uri, newPath := rule.FindPath("/api/test/ping")

	t.Log(uri, newPath)
}
