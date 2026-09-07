package config

import "testing"

func TestFromEnvDefaults(t *testing.T) {
	t.Setenv("GALLERY_LISTEN_ADDR", "")
	t.Setenv("GALLERY_DATA_DIR", "")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected empty configuration to fail")
	}
	t.Setenv("GALLERY_LISTEN_ADDR", ":9090")
	t.Setenv("GALLERY_DATA_DIR", "/var/lib/gallery")
	c, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if c.ListenAddr != ":9090" || c.DataDir != "/var/lib/gallery" {
		t.Fatalf("unexpected config: %+v", c)
	}
}
