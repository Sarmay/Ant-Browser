package browser

import (
	"ant-chrome/backend/internal/database"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSQLiteProxyDAOReplaceAllRollsBackOnInsertFailure(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "proxies.db"))
	if err != nil {
		t.Fatalf("NewDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	dao := NewSQLiteProxyDAO(db.GetConn())
	oldProxies := []Proxy{
		{
			ProxyId: "old-primary", ProxyName: "old primary", ProxyConfig: "socks5://127.0.0.1:1080",
			PreferredKernel: "xray", DnsServers: "1.1.1.1", GroupName: "old-group", SortOrder: 2,
		},
		{
			ProxyId: "old-secondary", ProxyName: "old secondary", ProxyConfig: "http://127.0.0.1:8080",
			SourceID: "old-source", SourceURL: "https://example.com/proxies", SourceNamePrefix: "old-",
			SourceAutoRefresh: true, SourceRefreshIntervalM: 30, SourceLastRefreshAt: "2026-08-08T00:00:00Z", SortOrder: 7,
		},
	}
	if err := dao.ReplaceAll(oldProxies); err != nil {
		t.Fatalf("initial ReplaceAll returned error: %v", err)
	}

	before, err := dao.List()
	if err != nil {
		t.Fatalf("List before failing replacement returned error: %v", err)
	}

	if _, err := db.GetConn().Exec(`
		CREATE TRIGGER fail_second_proxy_insert
		BEFORE INSERT ON browser_proxies
		WHEN NEW.proxy_id = 'new-secondary'
		BEGIN
			SELECT RAISE(ABORT, 'forced proxy insert failure');
		END`); err != nil {
		t.Fatalf("creating failure trigger returned error: %v", err)
	}

	err = dao.ReplaceAll([]Proxy{
		{ProxyId: "new-primary", ProxyName: "new primary", ProxyConfig: "socks5://127.0.0.1:2080", SortOrder: 0},
		{ProxyId: "new-secondary", ProxyName: "new secondary", ProxyConfig: "http://127.0.0.1:9080", SortOrder: 1},
	})
	if err == nil {
		t.Fatal("ReplaceAll succeeded despite the second insert trigger failure")
	}

	after, err := dao.List()
	if err != nil {
		t.Fatalf("List after failing replacement returned error: %v", err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("proxies after failed ReplaceAll = %#v, want original %#v", after, before)
	}
}
