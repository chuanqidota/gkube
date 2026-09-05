package labels

import (
	"encoding/json"
	"testing"
)

// TestLabelFilterUnmarshal verifies JSON deserialization of LabelFilter.
func TestLabelFilterUnmarshal(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantKey string
		wantOp  string
		wantLen int
		wantErr bool
	}{
		{"equal", `{"key":"app","operator":"=","values":["nginx"]}`, "app", "=", 1, false},
		{"not equal", `{"key":"app","operator":"!=","values":["nginx"]}`, "app", "!=", 1, false},
		{"in", `{"key":"app","operator":"in","values":["a","b"]}`, "app", "in", 2, false},
		{"notin", `{"key":"app","operator":"notin","values":["a"]}`, "app", "notin", 1, false},
		{"missing key still unmarshals", `{"operator":"=","values":["a"]}`, "", "=", 1, false},
		{"empty values still unmarshals", `{"key":"app","operator":"=","values":[]}`, "app", "=", 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var f LabelFilter
			err := json.Unmarshal([]byte(tc.input), &f)
			if (err != nil) != tc.wantErr {
				t.Fatalf("unmarshal: got err=%v, wantErr=%v", err, tc.wantErr)
			}
			if f.Key != tc.wantKey {
				t.Errorf("Key = %q, want %q", f.Key, tc.wantKey)
			}
			if f.Operator != tc.wantOp {
				t.Errorf("Operator = %q, want %q", f.Operator, tc.wantOp)
			}
			if len(f.Values) != tc.wantLen {
				t.Errorf("len(Values) = %d, want %d", len(f.Values), tc.wantLen)
			}
		})
	}
}

func TestBuildLabelSelector(t *testing.T) {
	cases := []struct {
		name    string
		filters []LabelFilter
		want    string
		wantErr bool
	}{
		{"empty", nil, "", false},
		{"single equal", []LabelFilter{{Key: "app", Operator: "=", Values: []string{"nginx"}}}, "app=nginx", false},
		{"single not equal", []LabelFilter{{Key: "env", Operator: "!=", Values: []string{"dev"}}}, "env!=dev", false},
		{"single in", []LabelFilter{{Key: "app", Operator: "in", Values: []string{"a", "b"}}}, "app in (a,b)", false},
		{"single notin", []LabelFilter{{Key: "env", Operator: "notin", Values: []string{"dev", "test"}}}, "env notin (dev,test)", false},
		{"multiple", []LabelFilter{
			{Key: "app", Operator: "=", Values: []string{"nginx"}},
			{Key: "env", Operator: "in", Values: []string{"prod", "staging"}},
		}, "app=nginx,env in (prod,staging)", false},
		{"equal needs exactly 1 value", []LabelFilter{{Key: "app", Operator: "=", Values: []string{"a", "b"}}}, "", true},
		{"not equal needs exactly 1 value", []LabelFilter{{Key: "app", Operator: "!=", Values: []string{}}}, "", true},
		{"invalid operator", []LabelFilter{{Key: "app", Operator: "like", Values: []string{"a"}}}, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildLabelSelector(tc.filters)
			if (err != nil) != tc.wantErr {
				t.Errorf("BuildLabelSelector: got err=%v, wantErr=%v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("BuildLabelSelector: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildCacheKey(t *testing.T) {
	cases := []struct {
		name, cluster, ns, rt, want string
	}{
		{"namespaced", "prod", "default", "deployment", "prod/default/deployment"},
		{"cluster-scoped", "prod", "", "node", "prod/__cluster__/node"},
		{"empty ns", "dev", "", "pod", "dev/__cluster__/pod"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildCacheKey(tc.cluster, tc.ns, tc.rt)
			if got != tc.want {
				t.Errorf("buildCacheKey(%q,%q,%q) = %q, want %q", tc.cluster, tc.ns, tc.rt, got, tc.want)
			}
		})
	}
}

func TestGetGVR(t *testing.T) {
	cases := []struct {
		name        string
		resource    string
		wantErr     bool
		wantGroup   string
		wantVersion string
		wantRes     string
	}{
		{"deployment", "deployment", false, "apps", "v1", "deployments"},
		{"pod", "pod", false, "", "v1", "pods"},
		{"custom", "argoproj.io/v1alpha1/applications", false, "argoproj.io", "v1alpha1", "applications"},
		{"invalid", "badformat", true, "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gvr, err := getGVR(tc.resource)
			if (err != nil) != tc.wantErr {
				t.Errorf("getGVR(%q): got err=%v, wantErr=%v", tc.resource, err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if gvr.Group != tc.wantGroup || gvr.Version != tc.wantVersion || gvr.Resource != tc.wantRes {
					t.Errorf("getGVR(%q) = %+v, want {Group:%s Version:%s Resource:%s}", tc.resource, gvr, tc.wantGroup, tc.wantVersion, tc.wantRes)
				}
			}
		})
	}
}

func TestIsBuiltinResource(t *testing.T) {
	cases := []struct {
		resource string
		want     bool
	}{
		{"deployment", true},
		{"pod", true},
		{"customresourcedefinition", true},
		{"argoproj.io/v1alpha1/applications", false},
		{"unknown", false},
	}
	for _, tc := range cases {
		t.Run(tc.resource, func(t *testing.T) {
			got := isBuiltinResource(tc.resource)
			if got != tc.want {
				t.Errorf("isBuiltinResource(%q) = %v, want %v", tc.resource, got, tc.want)
			}
		})
	}
}
