package ingest

import (
	"context"
	"testing"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/data/types"
	"github.com/Azure/azure-kusto-go/kusto/ingest/internal/resources"
)

type mockClient struct {
	endpoint string
	auth     kusto.Authorization
}

func (m mockClient) Auth() kusto.Authorization {
	return m.auth
}

func (m mockClient) Endpoint() string {
	return m.endpoint
}

func (m mockClient) Query(context.Context, string, kusto.Stmt, ...kusto.QueryOption) (*kusto.RowIterator, error) {
	panic("not implemented")
}

func (m mockClient) Mgmt(context.Context, string, kusto.Stmt, ...kusto.MgmtOption) (*kusto.RowIterator, error) {
	rows, err := kusto.NewMockRows(table.Columns{
		{
			Name: "ResourceTypeName",
			Type: types.String,
		},
		{
			Name: "StorageRoot",
			Type: types.String,
		},
	})
	if err != nil {
		return nil, err
	}
	iter := &kusto.RowIterator{}
	err = iter.Mock(rows)
	if err != nil {
		return nil, err
	}

	return iter, nil
}

func TestManager(t *testing.T) {

	firstMockClient := mockClient{
		endpoint: "https://test.kusto.windows.net",
		auth:     kusto.Authorization{},
	}
	mockClientSame := mockClient{
		endpoint: "https://test.kusto.windows.net",
		auth:     kusto.Authorization{},
	}
	secondMockClient := mockClient{
		endpoint: "https://test2.kusto.windows.net",
		auth:     kusto.Authorization{},
	}

	tests := []struct {
		name    string
		clients []func() *Ingestion
	}{
		{
			name: "TestSameClient",
			clients: []func() *Ingestion{
				func() *Ingestion {
					ingestion, _ := New(firstMockClient, "test", "test")
					return ingestion
				},
				func() *Ingestion {
					ingestion, _ := New(firstMockClient, "test2", "test2")
					return ingestion
				},
			},
		},
		{
			name: "TestSameEndpoint",
			clients: []func() *Ingestion{
				func() *Ingestion {
					ingestion, _ := New(firstMockClient, "test", "test")
					return ingestion
				},
				func() *Ingestion {
					ingestion, _ := New(mockClientSame, "test2", "test2")
					return ingestion
				},
			},
		},
		{
			name: "TestDifferentEndpoint",
			clients: []func() *Ingestion{
				func() *Ingestion {
					ingestion, _ := New(firstMockClient, "test", "test")
					return ingestion
				},
				func() *Ingestion {
					ingestion, _ := New(secondMockClient, "test2", "test2")
					return ingestion
				},
			},
		},
		{
			name: "TestDifferentAndSameEndpoints",
			clients: []func() *Ingestion{
				func() *Ingestion {
					ingestion, _ := New(firstMockClient, "test", "test")
					return ingestion
				},
				func() *Ingestion {
					ingestion, _ := New(mockClientSame, "test", "test")
					return ingestion
				},
				func() *Ingestion {
					ingestion, _ := New(secondMockClient, "test2", "test2")
					return ingestion
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mgrMap := make(map[*resources.Manager]bool)
			for _, client := range test.clients {
				mgr := client().mgr
				mgrMap[mgr] = true
			}

			if len(mgrMap) != len(test.clients) {
				t.Errorf("Got duplicated managers, want %d managers, got %d", len(test.clients), len(mgrMap))
			}
		})
	}
}
