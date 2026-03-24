package service

import "testing"

type keywordStub struct {
	Name      string
	Namespace string
	Extra     string
}

func TestFilterByKeywordFields_BoundaryAndCrossField(t *testing.T) {
	items := []keywordStub{
		{Name: "gateway-svc", Namespace: "prod", Extra: "10.0.0.1"},
		{Name: "payment-svc", Namespace: "staging", Extra: "10.0.0.2"},
	}

	shortKeyword := filterByKeywordFields(items, "ga", func(item keywordStub) []string {
		return []string{item.Name, item.Namespace, item.Extra}
	})
	if len(shortKeyword) != 2 {
		t.Fatalf("expected short keyword ignored, got %d", len(shortKeyword))
	}

	namespaceKeyword := filterByKeywordFields(items, "staging", func(item keywordStub) []string {
		return []string{item.Name, item.Namespace, item.Extra}
	})
	if len(namespaceKeyword) != 1 || namespaceKeyword[0].Name != "payment-svc" {
		t.Fatalf("expected namespace field matched payment-svc, got %d", len(namespaceKeyword))
	}
}

func TestFlattenLabels_StableOutput(t *testing.T) {
	labels := map[string]string{"team": "ops", "env": "prod"}
	output := flattenLabels(labels)
	if output != "env=prod,team=ops" {
		t.Fatalf("unexpected flatten labels output: %s", output)
	}
}
